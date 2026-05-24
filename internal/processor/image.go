package processor

import (
	"bytes"
	"context" // Add this
	_ "image/jpeg"
	"io"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/evanoberholster/imagemeta"
)

// S3Service updated for v2 signatures
type S3Service interface {
	GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
	CopyObject(ctx context.Context, params *s3.CopyObjectInput, optFns ...func(*s3.Options)) (*s3.CopyObjectOutput, error)
	DeleteObject(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error)
	// ... add others if needed
}

// GetImage retrieves the byte contents of a specified reader
func GetImage(r io.Reader) (*[]byte, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return &[]byte{}, err
	}
	return &data, nil
}

func GetImageReader(ctx context.Context, service S3Service, bucket string, key string) (io.Reader, error) {
	result, err := service.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	})
	if err != nil {
		return nil, err
	}
	return result.Body, nil
}

// GetImgTimeStamp tries to get the EXIF timestamp for the image the supplied reader refers to.
// it will return an error and nil Time if the object cannot be retrieved. If there are
// problems obtaining a meaningful timestamp from the file, it will return the current time.
func GetImgTimeStamp(image *[]byte) (*time.Time, error) {

	metaData, err := imagemeta.Decode(bytes.NewReader(*image))
	if err != nil {
		log.Printf("Failed to get metadata from image file: %v", err)
		return nil, err
	}

	if !metaData.OriginalDate().IsZero() {
		t := metaData.OriginalDate()
		return &t, nil
	}

	if !metaData.DigitizedDate().IsZero() {
		t := metaData.DigitizedDate()
		return &t, nil
	}

	if !metaData.ModifyDate().IsZero() {
		t := metaData.ModifyDate()
		return &t, nil
	}

	t := time.Now()
	return &t, nil
}
