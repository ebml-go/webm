package webm

import (
	"code.google.com/p/ebml-go/ebml"
	"io"
	"log"
	"time"
)

const (
	BadTC    = time.Duration(-1000000000000000)
	shutdown = 2 * BadTC
)

type Packet struct {
	Data        []byte
	Timecode    time.Duration
	TrackNumber uint
	Invisible   bool
	Keyframe    bool
	Discardable bool
	Rebase      bool
}

type Reader struct {
	Chan   chan Packet
	seek   chan time.Duration
	index  seekIndex
	offset int64
}

func (r *Reader) send(p *Packet) {
	r.Chan <- *p
}

func remaining(x int8) (rem int) {
	for x > 0 {
		rem++
		x += x
	}
	return
}

func laceSize(v []byte) (val int, rem int) {
	val = int(v[0])
	rem = remaining(int8(val))
	for i, l := 1, rem+1; i < l; i++ {
		val <<= 8
		val += int(v[i])
	}
	val &= ^(128 << uint(rem*8-rem))
	return
}

func laceDelta(v []byte) (val int, rem int) {
	val, rem = laceSize(v)
	val -= (1 << (uint(7*(rem+1) - 1))) - 1
	return
}

func (r *Reader) sendLaces(p *Packet, d []byte, sz []int) {
	var curr int
	for i, l := 0, len(sz); i < l; i++ {
		if sz[i] != 0 {
			p.Data = d[curr : curr+sz[i]]
			r.send(p)
			curr += sz[i]
			p.Timecode = BadTC
		}
	}
	p.Data = d[curr:]
	r.send(p)
}

func parseXiphSizes(d []byte) (sz []int, curr int) {
	laces := int(uint(d[4]))
	sz = make([]int, laces)
	curr = 5
	for i := 0; i < laces; i++ {
		for d[curr] == 255 {
			sz[i] += 255
			curr++
		}
		sz[i] += int(uint(d[curr]))
		curr++
	}
	return
}

func parseFixedSizes(d []byte) (sz []int, curr int) {
	laces := int(uint(d[4]))
	curr = 5
	fsz := len(d[curr:]) / (laces + 1)
	sz = make([]int, laces)
	for i := 0; i < laces; i++ {
		sz[i] = fsz
	}
	return
}

func parseEBMLSizes(d []byte) (sz []int, curr int) {
	laces := int(uint(d[4]))
	sz = make([]int, laces)
	curr = 5
	var rem int
	sz[0], rem = laceSize(d[curr:])
	for i := 1; i < laces; i++ {
		curr += rem + 1
		var dsz int
		dsz, rem = laceDelta(d[curr:])
		sz[i] = sz[i-1] + dsz
	}
	curr += rem + 1
	return
}

func (r *Reader) sendBlock(data []byte, tbase time.Duration) {
	var p Packet
	p.TrackNumber = uint(data[0]) & 0x7f
	p.Timecode = tbase + time.Millisecond*time.Duration(
		uint(data[1])<<8+uint(data[2]))
	p.Invisible = (data[3] & 8) != 0
	p.Keyframe = (data[3] & 0x80) != 0
	p.Discardable = (data[3] & 1) != 0
	if p.Discardable {
		log.Println("Discardable packet")
	}
	lacing := (data[3] >> 1) & 3
	switch lacing {
	case 0:
		p.Data = data[4:]
		r.send(&p)
	case 1:
		sz, curr := parseXiphSizes(data)
		r.sendLaces(&p, data[curr:], sz)
	case 2:
		sz, curr := parseFixedSizes(data)
		r.sendLaces(&p, data[curr:], sz)
	case 3:
		sz, curr := parseEBMLSizes(data)
		r.sendLaces(&p, data[curr:], sz)
	}
}

func (r *Reader) sendCluster(elmts *ebml.Element, tbase time.Duration) {
	var err error
	for err == nil && len(r.seek) == 0 {
		var e *ebml.Element
		e, err = elmts.Next()
		var blk []byte
		if err == nil {
			switch e.Id {
			case 0xa3:
				if err == nil {
					blk, err = e.ReadData()
				}
				if err != nil && err != io.EOF {
					log.Println(err)
				}
			case 0xa0:
				var bg BlockGroup
				err = e.Unmarshal(&bg)
				if err == nil {
					blk = bg.Block
				}
				if err != nil && err != io.EOF {
					log.Println(err)
				}
			default:
				log.Printf("Unexpected packet %x", e.Id)
			}

			if err == nil && blk != nil && len(blk) > 4 {
				r.sendBlock(blk, tbase)
			}
		}
	}
}

func (r *Reader) parseClusters(elmts *ebml.Element) {
	var err error
	for err == nil {
		var c Cluster
		var e *ebml.Element
		e, err = elmts.Next()
		if err == nil {
			err = e.Unmarshal(&c)
		}
		if err != nil && err.Error() == "Reached payload" {
			r.index.append(seekEntry{time.Millisecond * time.Duration(c.Timecode), e.Offset})
			r.sendCluster(err.(ebml.ReachedPayloadError).Element,
				time.Millisecond*time.Duration(c.Timecode))
			err = nil
		}
		seek := BadTC
		for len(r.seek) != 0 {
			seek = <-r.seek
		}
		if err == io.EOF {
			var eofpkt Packet
			eofpkt.Timecode = BadTC
			r.send(&eofpkt)
			seek = <-r.seek
			if seek != BadTC {
				err = nil
			}
		}
		if seek != BadTC {
			entry := r.index.search(seek)
			elmts.Seek(entry.offset, 0)
			var seekpkt Packet
			seekpkt.Timecode = seek
			r.send(&seekpkt)
		}
		if seek == shutdown {
			err = io.EOF
		}
	}
	close(r.Chan)
}

func newReader(e *ebml.Element, cuepoints []CuePoint, offset int64) *Reader {
	r := &Reader{make(chan Packet, 4), make(chan time.Duration, 4),
		*newSeekIndex(), offset}
	for i, l := 0, len(cuepoints); i < l; i++ {
		c := cuepoints[i]
		r.index.append(seekEntry{
			time.Millisecond * time.Duration(c.CueTime),
			offset + c.CueTrackPositions[0].CueClusterPosition,
		})
	}
	go r.parseClusters(e)
	return r
}

func (r *Reader) Seek(t time.Duration) {
	r.seek <- t
}

func (r *Reader) Shutdown() {
	r.seek <- shutdown
}
