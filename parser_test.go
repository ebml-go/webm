package webm

import (
	"log"
	"os"
	"testing"
)

func TestReadStruct(t *testing.T) {
	// sample fetch from http://techslides.com/sample-webm-ogg-and-mp4-video-files-for-html5
	// TODO: use much smaller one instead
	path := "testdata/small.webm"
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
