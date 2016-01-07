package webm

import (
	"fmt"
	"log"
	"time"

	"github.com/petar/GoLLRB/llrb"
)

type seekEntry struct {
	dur    time.Duration
	offset int64
}

// Less implements llrb.Item
func (it seekEntry) Less(it2 llrb.Item) bool {
	return it.dur < it2.(seekEntry).dur
}

func (it seekEntry) String() string {
	return fmt.Sprintf("{%v %v}", it.dur, it.offset)
}

type seekIndex struct {
	Tree *llrb.LLRB
}

func newSeekIndex() seekIndex {
	return seekIndex{
		Tree: llrb.New(),
	}
}

func (idx seekIndex) append(item seekEntry) {
	prev := idx.search(item.dur)
	if false && prev.dur != item.dur {
		log.Println("New entry", item)
	}
	if false && prev.dur == item.dur && prev.offset != item.offset {
		log.Println("Overriding entry", prev, item)
	}
	idx.Tree.ReplaceOrInsert(item)
}

func (idx seekIndex) search(dur time.Duration) (result seekEntry) {
	idx.Tree.AscendGreaterOrEqual(seekEntry{dur, 0}, func(item llrb.Item) bool {
		result = item.(seekEntry)
		return false
	})
	return
}
