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
	PhotoPrefix       string
	DestinationPrefix string
	DestinationBucket string
	Region            string
	S3service         *s3.S3
}

var params *runtimeParameters

const (
	DefaultPrefix = "photos/"
	DefaultRegion = "eu-west-2"
	DefaultBucket = "NOSUCHBUCKET"
)

func init() {
	params = &runtimeParameters{
		PhotoPrefix:       validatePrefix(os.Getenv("PHOTO_PREFIX")),
		DestinationPrefix: validatePrefix(os.Getenv("DESTINATION_PREFIX")),
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
		if strings.HasPrefix(event.S3.Object.Key, params.PhotoPrefix) && strings.HasPrefix(event.EventName, "ObjectCreated:") {
			decodedKey, err := url.QueryUnescape(event.S3.Object.Key)
			if err != nil {
				log.Printf("Failed to decode the key: '%s'", event.S3.Object.Key)
				break
			}

			if event.AWSRegion != params.Region {
				log.Printf("Event is not from the same region as the lambda: got %q, wanted %q", event.AWSRegion, params.Region)
				break
			}

			tstamp, err := getImgTimeStamp(event.S3.Bucket.Name, decodedKey)
			if err != nil {
				log.Printf("failed to obtain timestamp: %v", err)
				break
			}

			newKey := makeNewKey(decodedKey, tstamp)
			log.Printf("Processing request for : object %s/%s -> %s", event.S3.Bucket.Name, decodedKey, newKey)

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