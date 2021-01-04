package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"io"
	"net/url"
	"path/filepath"
	"strings"
	"time"
)

// s3Service helps with mocking access to S3
type s3Service interface {
	GetObject(input *s3.GetObjectInput) (*s3.GetObjectOutput, error)
	CopyObject(input *s3.CopyObjectInput) (*s3.CopyObjectOutput, error)
	WaitUntilObjectExists(input *s3.HeadObjectInput) error
	DeleteObject(input *s3.DeleteObjectInput) (*s3.DeleteObjectOutput, error)
}

// extractName gets the last part of the S3 key
func extractName(key string) string {
	if key == "" || strings.HasSuffix(key, "/") {
		return ""
	}
	return filepath.Base(key)
}

// makeNewKey will assemble the target key for a provided incoming object key, and the timestamp
func makeNewKey(key string, tstamp *time.Time) string {
	return params.DestinationPrefix + tstamp.Format("2006/01/02/") + extractName(key)
}

// getImageReader tries to get an io.Reader exposing the body of an image given the bucket and key. It will fail
// if the provided object is not a JPEG
func getImageReader(service s3Service, bucket string, key string) (io.Reader, error) {
	result, err := service.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("error fetchign from s3: %v", err)
	}

	if *result.ContentType != JPEG {
		return nil, fmt.Errorf("only JPEG supported, fetched file was reported as %s", *result.ContentType)
	}

	return result.Body, nil
}

func moveObject(service s3Service, srcBucket string, srcKey string, destBucket string, destKey string) error {
	// copy the object to the new location
	_, err := service.CopyObject(&s3.CopyObjectInput{
		Bucket: aws.String(destBucket),
		Key: aws.String(destKey),
		CopySource: aws.String(url.PathEscape(srcBucket+ "/" + srcKey)),
	})
	if err != nil {
		return fmt.Errorf("failed to copy object to destination: %v", err)
	}

	// verify it is there. looking at the source code, this has a comfortable retry and wait behaviour
	err = service.WaitUntilObjectExists(&s3.HeadObjectInput{
		Bucket:               aws.String(destBucket),
		Key:                  aws.String(destKey),
	})
	if err != nil {
		return fmt.Errorf("object was not available in the bucket after copying: %v", err)
	}

	// delete the original object
	_, err = service.DeleteObject(&s3.DeleteObjectInput{
		Bucket:                    aws.String(srcBucket),
		Key:                       aws.String(srcKey),
	})
	if err != nil {
		return fmt.Errorf("failed to delete original object after copying: %v", err)
	}

	return nil
}