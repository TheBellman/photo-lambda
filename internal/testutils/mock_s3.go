package testutils

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// MockS3 must satisfy processor.S3Service using SDK v2 signatures
type MockS3 struct {
	// Update this signature
	TestDataReader func(name string) (io.ReadCloser, error)
}

func (m *MockS3) GetObject(ctx context.Context, input *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	key := *input.Key
	var mime string
	var fileName string

	switch {
	case strings.Contains(key, "good.jpeg") || strings.Contains(key, "test.jpeg"):
		mime = "image/jpeg"
		fileName = "test.jpeg"
	case strings.Contains(key, "test.CR3"):
		mime = "binary/octet-stream"
		fileName = "test.CR3"
	case strings.Contains(key, "test.HEIC"):
		mime = "image/heic"
		fileName = "test.HEIC"
	case strings.Contains(key, "bad.jpeg"):
		mime = "text/plain"
		fileName = "test.jpeg"
	default:
		return nil, fmt.Errorf("unexpected test key provided: %s", key)
	}

	body, err := m.TestDataReader(fileName)
	if err != nil {
		return nil, fmt.Errorf("mock failed to load file: %v", err)
	}

	return &s3.GetObjectOutput{
		ContentType: &mime,
		Body:        body,
	}, nil
}

func (m *MockS3) CopyObject(ctx context.Context, input *s3.CopyObjectInput, optFns ...func(*s3.Options)) (*s3.CopyObjectOutput, error) {
	if strings.Contains(*input.Key, "case1") {
		return nil, fmt.Errorf("simulated copy failure")
	}
	return &s3.CopyObjectOutput{}, nil
}

func (m *MockS3) DeleteObject(ctx context.Context, input *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error) {
	if strings.Contains(*input.Key, "case3") {
		return nil, fmt.Errorf("simulated delete failure")
	}
	return &s3.DeleteObjectOutput{}, nil
}

// Ensure these match whatever is in your S3Service interface
func (m *MockS3) PutObject(ctx context.Context, input *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	return &s3.PutObjectOutput{}, nil
}

// Note: v2 SDK doesn't have WaitUntilObjectExists on the client.
// If your S3Service interface still includes it, you'll need to decide
// whether to keep it or use the v2 Waiter pattern.
