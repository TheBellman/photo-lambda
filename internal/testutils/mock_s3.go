package testutils

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/service/s3"
)

type MockS3 struct {
	TestDataReader func(name string) io.ReadCloser
}

func (m *MockS3) GetObject(input *s3.GetObjectInput) (*s3.GetObjectOutput, error) {
	key := *input.Key
	var mime string
	var filePath string

	switch {
	case strings.Contains(key, "good.jpeg") || strings.Contains(key, "test.jpeg"):
		mime = "image/jpeg"
		filePath = filepath.Join("..", "..", "testdata", "test.jpeg")
	case strings.Contains(key, "test.CR3"):
		mime = "binary/octet-stream"
		filePath = filepath.Join("..", "..", "testdata", "test.CR3")
	case strings.Contains(key, "test.HEIC"):
		mime = "image/heic"
		filePath = filepath.Join("..", "..", "testdata", "test.HEIC")
	case strings.Contains(key, "bad.jpeg"):
		mime = "text/plain"
		filePath = filepath.Join("..", "..", "testdata", "test.jpeg")
	default:
		return nil, fmt.Errorf("unexpected test key: %s", key)
	}

	return &s3.GetObjectOutput{
		ContentType: &mime,
		Body:        m.TestDataReader(filePath),
	}, nil
}

func (m *MockS3) CopyObject(input *s3.CopyObjectInput) (*s3.CopyObjectOutput, error) {
	if strings.Contains(*input.Key, "case1") {
		return nil, fmt.Errorf("copy failed")
	}
	return &s3.CopyObjectOutput{}, nil
}

func (m *MockS3) WaitUntilObjectExists(input *s3.HeadObjectInput) error {
	if strings.Contains(*input.Key, "case2") {
		return fmt.Errorf("wait failed")
	}
	return nil
}

func (m *MockS3) DeleteObject(input *s3.DeleteObjectInput) (*s3.DeleteObjectOutput, error) {
	if strings.Contains(*input.Key, "case3") {
		return nil, fmt.Errorf("delete failed")
	}
	return &s3.DeleteObjectOutput{}, nil
}

func (m *MockS3) PutObject(input *s3.PutObjectInput) (*s3.PutObjectOutput, error) {
	return &s3.PutObjectOutput{}, nil
}
