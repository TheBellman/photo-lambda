package main

import (
	"bytes"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/evanoberholster/imagemeta"
	_ "image/jpeg"
	"io"
	"log"
	"strings"
	"time"
)

// getImage retrieves the byte contents of a specified reader
func getImage(r io.Reader) (*[]byte, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return &[]byte{}, err
	}
	return &data, nil
}

// getImageReader tries to get an io.Reader exposing the body of an image given the bucket and key. It will fail
// if the provided object is not a supported file type
func getImageReader(service s3Service, bucket string, key string) (io.Reader, error) {
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
	return nil, fmt.Errorf("only JPEG and CR3 supported, fetched file %s was reported as %s",
		key,
		*result.ContentType)
}

// getImgTimeStamp tries to get the EXIF timestamp for the image the supplied reader refers to.
// it will return an error and nil Time if the object cannot be retrieved. If there are
// problems obtaining a meaningful timestamp from the file, it will return the current time.
func getImgTimeStamp(image *[]byte) (*time.Time, error) {

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
