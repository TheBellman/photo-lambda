package main

import (
	"io"
	"log"
	"os"
	"testing"
)

func testFileReader() io.ReadCloser {
	f, err := os.Open("./test.jpeg")
	if err != nil {
		log.Fatal("Failed to open test jpeg")
	}
	return f
}

func Test_getImgTimeStamp(t *testing.T) {
	f := testFileReader()
	tstamp, err := getImgTimeStamp(f)
	if err != nil {
		t.Errorf("Failed to extract timestamp: %v", err)
	}
	if tstamp == nil {
		t.Errorf("Received a nil timestamp")
	}
}


