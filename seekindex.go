package webm

import (
	"fmt"
	"github.com/petar/GoLLRB/llrb"
	"log"
	"time"
)

type seekEntry struct {
	t      time.Duration
	offset int64
}

func (se seekEntry) String() string {
	return fmt.Sprintf("{%v %v}", se.t, se.offset)
}

type seekIndex struct {
	t llrb.Tree
}

func newSeekIndex() *seekIndex {
	return &seekIndex{*llrb.New(func(a, b interface{}) bool {
		return a.(seekEntry).t > b.(seekEntry).t
	})}
}

func (si *seekIndex) append(se seekEntry) {
	prev := si.search(se.t)
	if false && prev.t != se.t {
		log.Println("New entry", se)
	}
	if false && prev.t == se.t && prev.offset != se.offset {
		log.Println("Overriding entry", prev, se)
	}
	si.t.ReplaceOrInsert(se)
}

func (si *seekIndex) search(t time.Duration) (val seekEntry) {
	si.t.AscendGreaterOrEqual(seekEntry{t, 0}, func(i llrb.Item) bool {
		val = i.(seekEntry)
		return false
	})
	return
}
