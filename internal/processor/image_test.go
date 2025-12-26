package processor

import (
	"io"
	"log"
	"os"
	"testing"

	"github.com/TheBellman/photo-lambda/internal/testutils"
)

func testFile(name string) *[]byte {
	data, err := io.ReadAll(testFileReader(name))
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
	keys := []string{"../../testdata/test.jpeg", "../../testdata/test.CR3", "../../testdata/test.HEIC"}
	for _, key := range keys {
		data, err := GetImage(testFileReader(key))
		if err != nil {
			t.Errorf("unexpected error loading file: %v", err)
		}
		if len(*data) == 0 {
			t.Errorf("empty byte slice returned!")
		}
	}
}

func Test_getImgTimeStamp(t *testing.T) {
	keys := []string{"../../testdata/test.jpeg", "../../testdata/test.CR3", "../../testdata/test.HEIC"}
	for _, key := range keys {
		tstamp, err := GetImgTimeStamp(testFile(key))
		if err != nil {
			t.Errorf("Failed to extract timestamp: %v", err)
		}
		if tstamp == nil {
			t.Errorf("Received a nil timestamp")
		}

	}
}

func Test_getImageReader(t *testing.T) {
	// Pass the local testFileReader to the mock
	mock := testutils.MockS3{TestDataReader: testFileReader}

	keys := []string{"key/good.jpeg", "key/test.CR3", "key/test.HEIC"}

	for _, key := range keys {
		_, err := GetImageReader(&mock, "bucket", key)
		if err != nil {
			t.Errorf("Received an unexpected error: %v", err)
		}

	}

	_, err := GetImageReader(&mock, "bucket", "key/bad.jpeg")
	if err == nil {
		t.Errorf("Did not get an error when expected")
	}
}
