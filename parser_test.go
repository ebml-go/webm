package webm

import (
	"os"
	"testing"

//	"code.google.com/p/ebml-go/ebml"
)

func TestReadStruct(t *testing.T) {
	path := "/Users/jacereda/Movies/big-buck-bunny_trailer.webm"
	r, err := os.Open(path)
	if err != nil {
		t.Fatal("unable to open file " + path)
	}
	var w WebM
	_, err = Parse(r, &w)
	t.Log("Duration: ", w.Segment.GetDuration())
	/*	for err == nil {
			var e *ebml.Element
			e, err = rest.Next()
			_, err = e.ReadData()
			t.Log("Packet: ", e.Id, e.Size())
		}
	*/
}
