package main

import (
	"io"
	"io/ioutil"
	"log"
	"os"
	"testing"
)

func testFile(name string) *[]byte {
	data, err := ioutil.ReadAll(testFileReader(name))
	if err != nil {
		log.Fatalf("Failed to read %s", name)
	}
	return &data
}

func testFileReader(name string) io.ReadCloser {
	f, err := os.Open(name)
	if err != nil {
		log.Fatalf("Failed to open %s", name)
	}
	return f
}

func Test_getImage(t *testing.T) {
	data, err := getImage(testFileReader("./test.jpeg"))
	if err != nil {
		t.Errorf("unexpected error loading file: %v", err)
	}
	if len(*data) == 0 {
		t.Errorf("empty byte slice returned!")
	}
}

func Test_getImgTimeStamp(t *testing.T) {
	tstamp, err := getImgTimeStamp(testFile("./test.jpeg"))
	if err != nil {
		t.Errorf("Failed to extract timestamp: %v", err)
	}
	if tstamp == nil {
		t.Errorf("Received a nil timestamp")
	}
}
