package processor

import (
	"bytes"
	"fmt"
	_ "image/jpeg"
	"io"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/evanoberholster/imagemeta"
)

const (
	JPEG = "image/jpeg"
	HEIC = "image/heic"
)

type S3Service interface {
	GetObject(input *s3.GetObjectInput) (*s3.GetObjectOutput, error)
	CopyObject(input *s3.CopyObjectInput) (*s3.CopyObjectOutput, error)
	WaitUntilObjectExists(input *s3.HeadObjectInput) error
	DeleteObject(input *s3.DeleteObjectInput) (*s3.DeleteObjectOutput, error)
	PutObject(input *s3.PutObjectInput) (*s3.PutObjectOutput, error)
}

// GetImage retrieves the byte contents of a specified reader
func GetImage(r io.Reader) (*[]byte, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return &[]byte{}, err
	}
	return &data, nil
}

// GetImageReader tries to get an io.Reader exposing the body of an image given the bucket and key. It will fail
// if the provided object is not a supported file type
func GetImageReader(service S3Service, bucket string, key string) (io.Reader, error) {
	result, err := service.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("error fetching from s3: %v", err)
	}

	if strings.HasSuffix(strings.ToLower(key), ".cr3") ||
		strings.HasSuffix(strings.ToLower(key), ".heic") ||
		*result.ContentType == HEIC ||
		*result.ContentType == JPEG {
		return result.Body, nil
	}
	return nil, fmt.Errorf("only JPEG, CR3  and HEIC supported, fetched file %s was reported as %s",
		key,
		*result.ContentType)
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

	if !metaData.DateTimeOriginal().IsZero() {
		t := metaData.DateTimeOriginal()
		return &t, nil
	}

	if !metaData.CreateDate().IsZero() {
		t := metaData.CreateDate()
		return &t, nil
	}

	if !metaData.ModifyDate().IsZero() {
		t := metaData.ModifyDate()
		return &t, nil
	}

	t := time.Now()
	return &t, nil
}
