package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/rwcarlsen/goexif/exif"
	"github.com/rwcarlsen/goexif/tiff"
	"io"
	"log"
	"time"
)

// getImgTimeStamp tries to get the EXIF timestamp for the specified object in the specified bucket.
// it will return an error and nil Time if the object cannot be retrieved, or is not a JPEG. If there are
// problems obtaining a meaningful timestamp from the JPEG, it will return the current time.
// ToDo: needs test
func getImgTimeStamp(bucket string, key string) (*time.Time, error) {

	r, err := getImageReader(bucket, key)
	if err != nil {
		return nil, err
	}

	metaData, err := exif.Decode(r)
	if err != nil {
		log.Printf("Failed to get metadata from image file: %v", err)
		return nil, err
	}

	var imgDate *tiff.Tag

	imgDate, err = metaData.Get("DateTimeOriginal")
	if err != nil {
		imgDate, err = metaData.Get("DateTimeDigitized")
		if err != nil {
			imgDate, err = metaData.Get("DateTime")
		}
	}

	if imgDate != nil {
		t, err := time.Parse("\"2006:01:02 15:04:05\"", imgDate.String())
		if err != nil {
			log.Printf("Failed to parse imgDate: %v", err)
			t = time.Now()
			return &t, nil
		}
		return &t, nil
	} else {
		t := time.Now()
		return &t, nil
	}
}

// getImageReader tries to get an io.Reader exposing the body of an image given the bucket and key. It will fail
// if the provided object is not a JPEG
func getImageReader(bucket string, key string) (io.Reader, error) {
	result, err := params.S3service.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		log.Printf("Error fetching from S3: %v", err)
		return nil, err
	}

	if *result.ContentType != "image/jpeg" {
		log.Printf("Only JPEG supported, fetched file was reported as %s", *result.ContentType)
		return nil, err
	}

	return result.Body, nil
}
