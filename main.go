// main provides a Lambda function used to archive and manipulate the photo stream
package main

import (
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"io"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type runtimeParameters struct {
	SourcePrefix      string
	DestinationPrefix string
	DestinationBucket string
	Region            string
	Session           *session.Session
	S3service         *s3.S3
}

// s3Service helps with mocking access to S3
type s3Service interface {
	GetObject(input *s3.GetObjectInput) (*s3.GetObjectOutput, error)
	CopyObject(input *s3.CopyObjectInput) (*s3.CopyObjectOutput, error)
	WaitUntilObjectExists(input *s3.HeadObjectInput) error
	DeleteObject(input *s3.DeleteObjectInput) (*s3.DeleteObjectOutput, error)
	PutObject(input *s3.PutObjectInput) (*s3.PutObjectOutput, error)
}

var params *runtimeParameters
var buildStamp string

const (
	DefaultRegion      = "eu-west-2"
	DefaultSrcPrefix   = "import/"
	DefaultDestPrefix  = "photos/"
	DefaultBucket      = "NOSUCHBUCKET"
	JPEG               = "image/jpeg"
)

func init() {
	buildStamp = os.Getenv("BUILD_STAMP")
	params = &runtimeParameters{
		SourcePrefix:      validatePrefix(os.Getenv("SOURCE_PREFIX"), DefaultSrcPrefix),
		DestinationPrefix: validatePrefix(os.Getenv("DESTINATION_PREFIX"), DefaultDestPrefix),
		DestinationBucket: validateDestination(os.Getenv("DESTINATION_BUCKET")),
		Region:            validateRegion(os.Getenv("AWS_REGION")),
	}
}

// validateRegion will provide the default region if no region is set
func validateRegion(region string) string {
	if region == "" {
		return DefaultRegion
	} else {
		return region
	}
}

// validatePrefix coerces the environmental variable into a usable prefix, by adding a "/" if necessary or setting it to
// the default prefix. It returns the coerced prefix
func validatePrefix(photoPrefix string, defaultPrefix string) string {
	if !strings.HasSuffix(photoPrefix, "/") {
		if photoPrefix == "" {
			photoPrefix = defaultPrefix
		} else {
			photoPrefix += "/"
		}
	}
	return photoPrefix
}

// validateDestination will ensure a non-blank destination bucket
func validateDestination(bucket string) string {
	if bucket == "" {
		return DefaultBucket
	} else {
		return bucket
	}
}

// makeNewKey will assemble the target key for a provided incoming object key, and the timestamp
func makeNewKey(key string, tstamp *time.Time) string {
	var dir, name = filepath.Split(key)
	dir = strings.TrimPrefix(dir, params.SourcePrefix)
	return params.DestinationPrefix + dir + tstamp.Format("2006/01/02/") + name
}

// getImageReader tries to get an io.Reader exposing the body of an image given the bucket and key. It will fail
// if the provided object is not a JPEG
func getImageReader(service s3Service, bucket string, key string) (io.Reader, error) {
	result, err := service.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("error fetching from s3: %v", err)
	}

	if *result.ContentType != JPEG {
		return nil, fmt.Errorf("only JPEG supported, fetched file was reported as %s", *result.ContentType)
	}

	return result.Body, nil
}

// moveObject uses the supplied service to move an object from a source bucket/key to a destination bucket/key
func moveObject(service s3Service, srcBucket string, srcKey string, destBucket string, destKey string) error {
	// silently do nothing if asked to move nowhere
	if srcBucket == destBucket && srcKey == destKey {
		return nil
	}

	// copy the object to the new location
	_, err := service.CopyObject(&s3.CopyObjectInput{
		Bucket:       aws.String(destBucket),
		Key:          aws.String(destKey),
		CopySource:   aws.String(url.PathEscape(fmt.Sprintf("%s/%s", srcBucket, srcKey))),
		StorageClass: aws.String("STANDARD_IA"),
	})
	if err != nil {
		return fmt.Errorf("failed to copy object to destination: %v", err)
	}

	// verify it is there. looking at the source code, this has a comfortable retry and wait behaviour
	err = service.WaitUntilObjectExists(&s3.HeadObjectInput{
		Bucket: aws.String(destBucket),
		Key:    aws.String(destKey),
	})
	if err != nil {
		return fmt.Errorf("object was not available in the bucket after copying: %v", err)
	}

	// delete the original object
	_, err = service.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(srcBucket),
		Key:    aws.String(srcKey),
	})
	if err != nil {
		return fmt.Errorf("failed to delete original object after copying: %v", err)
	}

	return nil
}

// HandleLambdaEvent takes care of processing the incoming S3 event. Only "ObjectCreated:*" events are processed, and only
// for where the object key starts with the nominated prefix. The count of processed objects is returned
func HandleLambdaEvent(request events.S3Event) (int, error) {
	cnt := 0
	for _, event := range request.Records {

		log.Printf("[%s] Received request for : object %s/%s", buildStamp, event.S3.Bucket.Name, event.S3.Object.Key)
		// only process events where the object key as the expected prefix and the event is an object creation
		if strings.HasPrefix(event.S3.Object.Key, params.SourcePrefix) && strings.HasPrefix(event.EventName, "ObjectCreated:") {
			decodedKey, err := url.QueryUnescape(event.S3.Object.Key)
			if err != nil {
				log.Printf("[%s] Failed to decode the key: '%s'", buildStamp, event.S3.Object.Key)
				continue
			}

			// this should be a cannot-happen case
			if event.AWSRegion != params.Region {
				log.Printf("[%s] Event is not from the same region as the lambda: got %q, wanted %q", buildStamp, event.AWSRegion, params.Region)
				continue
			}

			// fetch the object and hand back an io.reader
			imgReader, err := getImageReader(params.S3service, event.S3.Bucket.Name, decodedKey)
			if err != nil {
				log.Printf("[%s] Failed to get a reader to read from %s/%s: %v", buildStamp, event.S3.Bucket.Name, decodedKey, err)
				continue
			}

			// extract the image data
			imageBytes, err := getImage(imgReader)
			if err != nil {
				log.Printf("[%s] Failed to read image bytes: %v", buildStamp, err)
				continue
			}

			// try to get the EXIF timestamp for the object
			tstamp, err := getImgTimeStamp(imageBytes)
			if err != nil {
				log.Printf("[%s] failed to obtain timestamp: %v", buildStamp, err)
				continue
			}

			// use the EXIF timestamp and the supplied key to create a destination key
			newKey := makeNewKey(decodedKey, tstamp)

			// move the original object to it's new location
			if err = moveObject(params.S3service, event.S3.Bucket.Name, event.S3.Object.Key, params.DestinationBucket, newKey); err != nil {
				log.Printf("[%s] failed to move object: %v", buildStamp, err)
				continue
			}

			log.Printf("[%s] Processed request for : object %s/%s -> %s", buildStamp, event.S3.Bucket.Name, decodedKey, newKey)
			cnt++
		}
	}

	return cnt, nil
}

// main function invoked when the lambda is launched
func main() {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(params.Region),
	})
	if err != nil {
		log.Fatal("Error starting session", err)
	}
	params.Session = sess
	params.S3service = s3.New(sess)

	log.Printf("[%s] Registering handler for photo-lambda...", buildStamp)
	lambda.Start(HandleLambdaEvent)
}
