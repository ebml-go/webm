package webm

import (
	"log"
	"os"
	"testing"
)

func TestReadStruct(t *testing.T) {
	path := "video.webm"
	r, err := os.Open(path)
	if err != nil {
		t.Fatal("unable to open file", path)
	}
	var w WebM
	if _, err = Parse(r, &w); err != nil {
		t.Error()
	}
	log.Println("Duration:", w.Segment.GetDuration())
}
