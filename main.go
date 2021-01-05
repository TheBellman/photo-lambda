// main provides a Lambda function used to archive and manipulate the photo stream
package main

import (
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"log"
	"net/url"
	"os"
	"strings"
)

// photoPrefix contains the prefix we expect all keys to have if they are of interest
type runtimeParameters struct {
	SourcePrefix      string
	DestinationPrefix string
	DestinationBucket string
	Region            string
	S3service         *s3.S3
}

var params *runtimeParameters

const (
	DefaultRegion      = "eu-west-2"
	DefaultSrcPrefix   = "import/"
	DefaultDestPrefix  = "photos/"
	DefaultThumbPrefix = "thumbs/"
	DefaultBucket      = "NOSUCHBUCKET"
	JPEG               = "image/jpeg"
	ThumbnailSize      = 200
)

func init() {
	params = &runtimeParameters{
		SourcePrefix:      validatePrefix(os.Getenv("SOURCE_PREFIX"), DefaultSrcPrefix),
		DestinationPrefix: validatePrefix(os.Getenv("DESTINATION_PREFIX"), DefaultDestPrefix),
		DestinationBucket: validateDestination(os.Getenv("DESTINATION_BUCKET")),
		Region:            validateRegion(os.Getenv("AWS_REGION")),
	}

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(params.Region),
	})
	if err != nil {
		log.Fatal("Error starting session", err)
	}

	params.S3service = s3.New(sess)
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

// HandleLambdaEvent takes care of processing the incoming S3 event. Only "ObjectCreated:*" events are processed, and only
// for where the object key starts with the nominated prefix. The count of processed objects is returned
func HandleLambdaEvent(request events.S3Event) (int, error) {
	cnt := 0
	for _, event := range request.Records {
		// only process events where the object key as the expected prefix and the event is an object creation
		if strings.HasPrefix(event.S3.Object.Key, params.SourcePrefix) && strings.HasPrefix(event.EventName, "ObjectCreated:") {
			decodedKey, err := url.QueryUnescape(event.S3.Object.Key)
			if err != nil {
				log.Printf("Failed to decode the key: '%s'", event.S3.Object.Key)
				continue
			}

			// this should be a cannot-happen case
			if event.AWSRegion != params.Region {
				log.Printf("Event is not from the same region as the lambda: got %q, wanted %q", event.AWSRegion, params.Region)
				continue
			}

			// fetch the object and hand back an io.reader
			imgReader, err := getImageReader(params.S3service, event.S3.Bucket.Name, decodedKey)
			if err != nil {
				log.Printf("Failed to get a reader to read from %s/%s: %v", event.S3.Bucket.Name, decodedKey, err)
				continue
			}

			// extract the image data
			imageBytes, err := getImage(imgReader)
			if err != nil {
				log.Printf("Failed to read image bytes: %v", err)
			}

			// try to get the EXIF timestamp for the object
			tstamp, err := getImgTimeStamp(imageBytes)
			if err != nil {
				log.Printf("failed to obtain timestamp: %v", err)
				continue
			}

			// use the EXIF timestamp and the supplied key to create a destination key
			newKey := makeNewKey(decodedKey, tstamp)

			// move the original object to it's new location
			err = moveObject(params.S3service, event.S3.Bucket.Name, event.S3.Object.Key, params.DestinationBucket, newKey)
			if err != nil {
				log.Printf("failed to move object: %v", err)
				continue
			}

			// create a thumbnail from our image bytes, getting back a *byte[]
			thumbBytes, err := resizeImage(imageBytes)
			if err != nil {
				log.Printf("failed to create a thumbnail image: %v", err)
				continue
			}

			if err = saveThumbnail(params.S3service, thumbBytes, params.DestinationBucket, makeThumbKey(newKey)); err != nil {
				log.Printf("failed to save the thumbnail: %v", err)
			}

			log.Printf("Processed request for : object %s/%s -> %s", event.S3.Bucket.Name, decodedKey, newKey)
			cnt++
		}
	}

	return cnt, nil
}

// main function invoked when the lambda is launched
func main() {
	log.Println("Registering handler...")
	lambda.Start(HandleLambdaEvent)
}
