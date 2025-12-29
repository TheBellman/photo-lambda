package processor

import (
	"context" // Add this
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/TheBellman/photo-lambda/internal/testutils"
)

func testFile(name string) *[]byte {
	data, err := io.ReadAll(mustOpenFile(name))
	if err != nil {
		log.Fatalf("Failed to read %s", name)
	}
	return &data
}

func testFileReader(name string) (io.ReadCloser, error) {
	path := filepath.Join("..", "..", "testdata", name)
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return f, nil
}

// Updated helper for tests that expect the file to definitely exist
func mustOpenFile(name string) io.ReadCloser {
	f, err := testFileReader(name)
	if err != nil {
		log.Fatalf("Failed to open %s: %v", name, err)
	}
	return f
}

func Test_getImage(t *testing.T) {
	keys := []string{"test.jpeg", "test.CR3", "test.HEIC"}
	for _, key := range keys {
		data, err := GetImage(mustOpenFile(key))
		if err != nil {
			t.Errorf("unexpected error loading file: %v", err)
		}
		if len(*data) == 0 {
			t.Errorf("empty byte slice returned!")
		}
	}
}

func Test_getImgTimeStamp(t *testing.T) {
	keys := []string{"test.jpeg", "test.CR3", "test.HEIC"}
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
	// FIX: Use testFileReader here, because the mock expects the (io.ReadCloser, error) signature
	mock := testutils.MockS3{TestDataReader: testFileReader}
	ctx := context.TODO()

	keys := []string{"key/good.jpeg", "key/test.CR3", "key/test.HEIC"}

	for _, key := range keys {
		_, err := GetImageReader(ctx, &mock, "bucket", key)
		if err != nil {
			t.Errorf("Received an unexpected error: %v", err)
		}
	}
}

func Test_GetImageReader_FileNotFound(t *testing.T) {
	mock := testutils.MockS3{TestDataReader: testFileReader}
	ctx := context.TODO()

	// This key is NOT in the MockS3 switch, so it triggers the 'default' error
	_, err := GetImageReader(ctx, &mock, "bucket", "missing-file.jpg")

	if err == nil {
		t.Fatal("Expected an error for a non-existent file, but got nil")
	}

	expectedError := "unexpected test key provided: missing-file.jpg"
	if !strings.Contains(err.Error(), "unexpected test key") {
		t.Errorf("Expected error message to contain '%s', but got: %v", expectedError, err)
	}
}
