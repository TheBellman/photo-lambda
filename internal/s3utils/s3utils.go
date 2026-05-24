package s3utils

import (
	"context"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/TheBellman/photo-lambda/internal/processor"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Config struct {
	SourcePrefix      string
	DestinationPrefix string
	ErrorPrefix       string
}

// makeNewKey will assemble the target key for a provided incoming object key, and the timestamp
func makeNewKey(config Config, key string, tstamp *time.Time) string {
	var dir, name = filepath.Split(key)
	dir = strings.TrimPrefix(dir, config.SourcePrefix)
	return config.DestinationPrefix + dir + tstamp.Format("2006/01/02/") + name
}

// makeErrKey will take an input key and from it generate an error key
func makeErrKey(config Config, key string) string {
	return strings.Replace(filepath.Clean(key), config.SourcePrefix, config.ErrorPrefix, 1)
}

// MoveObject uses the supplied service to move an object
func MoveObject(ctx context.Context, service processor.S3Service, srcBucket string, srcKey string, destBucket string, destKey string) error {
	// silently do nothing if asked to move nowhere
	if srcBucket == destBucket && srcKey == destKey {
		return nil
	}

	// copy the object to the new location
	_, err := service.CopyObject(ctx, &s3.CopyObjectInput{
		Bucket:       aws.String(destBucket),
		Key:          aws.String(destKey),
		CopySource:   aws.String(url.PathEscape(fmt.Sprintf("%s/%s", srcBucket, srcKey))),
		StorageClass: "STANDARD_IA",
	})
	if err != nil {
		return fmt.Errorf("failed to copy object to destination: %v", err)
	}

	// delete the original object
	_, err = service.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(srcBucket),
		Key:    aws.String(srcKey),
	})
	if err != nil {
		return fmt.Errorf("failed to delete original object after copying: %v", err)
	}

	return nil
}

// ProcessAndMoveObject handles the logic of determining the new key and moving the object.
func ProcessAndMoveObject(ctx context.Context, service processor.S3Service, config Config, srcBucket string, srcKey string, decodedKey string, tstamp *time.Time, destBucket string) error {
	newKey := makeNewKey(config, decodedKey, tstamp)
	return MoveObject(ctx, service, srcBucket, srcKey, destBucket, newKey)
}

// SaveErrorObject tries to preserve an input object into an error folder
func SaveErrorObject(ctx context.Context, service processor.S3Service, config Config, bucket string, objectKey string, decodedKey string) error {
	errKey := makeErrKey(config, decodedKey)
	return MoveObject(ctx, service, bucket, objectKey, bucket, errKey)
}
