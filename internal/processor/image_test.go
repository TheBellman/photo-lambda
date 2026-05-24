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
	tests := []struct {
		key          string
		expectedTime string
	}{
		{"test.jpeg", "2006-06-20 00:46:21"},
		{"test.CR3", "2022-04-02 08:18:55"},
		{"test.HEIC", "2022-09-30 18:06:02"},
		{"test.ORF", "2024-06-09 13:52:20"},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			tstamp, err := GetImgTimeStamp(testFile(tt.key), tt.key)
			if err != nil {
				t.Errorf("Failed to extract timestamp for %s: %v", tt.key, err)
				return
			}
			if tstamp == nil {
				t.Errorf("Received a nil timestamp for %s", tt.key)
				return
			}

			actualTime := tstamp.Format("2006-01-02 15:04:05")
			if actualTime != tt.expectedTime {
				t.Errorf("For %s, expected timestamp %s, but got %s", tt.key, tt.expectedTime, actualTime)
			}
		})
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
