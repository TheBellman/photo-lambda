// main provides a Lambda function used to archive and manipulate the photo stream.
// This function will try to archive any file it's told about, so it relies on the
// notification configuration to filter for only certain files. It does however
// treat *.ORF files differently to other files, as the mechanism for digging
// the photo date out for those is different than HEIC, JPEG and CR3 files.
package main

import (
	"context"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/TheBellman/photo-lambda/internal/processor"
	"github.com/TheBellman/photo-lambda/internal/s3utils"
	"github.com/TheBellman/photo-lambda/internal/validate"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type runtimeParameters struct {
	SourcePrefix      string
	ErrorPrefix       string
	DestinationPrefix string
	DestinationBucket string
	Region            string
	S3Client          *s3.Client // Changed from *s3.S3
}

var params *runtimeParameters
var buildStamp string

const (
	DefaultRegion     = "eu-west-2"
	DefaultSrcPrefix  = "import/"
	DefaultErrPrefix  = "errors/"
	DefaultDestPrefix = "photos/"
	DefaultBucket     = "NOSUCHBUCKET"
)

func init() {
	buildStamp = os.Getenv("BUILD_STAMP")
	params = &runtimeParameters{
		SourcePrefix:      validate.Prefix(os.Getenv("SOURCE_PREFIX"), DefaultSrcPrefix),
		DestinationPrefix: validate.Prefix(os.Getenv("DESTINATION_PREFIX"), DefaultDestPrefix),
		ErrorPrefix:       validate.Prefix(os.Getenv("ERROR_PREFIX"), DefaultErrPrefix),
		DestinationBucket: validate.Destination(os.Getenv("DESTINATION_BUCKET"), DefaultBucket),
		Region:            validate.Region(os.Getenv("AWS_REGION"), DefaultRegion),
	}
}

// HandleLambdaEvent takes care of processing the incoming S3 event. Only "ObjectCreated:*" events are processed, and only
// for where the object key starts with the nominated prefix. The count of processed objects is returned
func HandleLambdaEvent(ctx context.Context, request events.S3Event) (int, error) {
	s3Config := s3utils.Config{
		SourcePrefix:      params.SourcePrefix,
		DestinationPrefix: params.DestinationPrefix,
		ErrorPrefix:       params.ErrorPrefix,
	}
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

			// Pass the ctx to your functions
			imgReader, err := processor.GetImageReader(ctx, params.S3Client, event.S3.Bucket.Name, event.S3.Object.Key)
			if err != nil {
				log.Printf("[%s] Failed to get a reader to read from %s/%s: %v", buildStamp, event.S3.Bucket.Name, decodedKey, err)
				continue
			}

			// extract the image data
			imageBytes, err := processor.GetImage(imgReader)
			if err != nil {
				log.Printf("[%s] Failed to read image bytes: %v", buildStamp, err)
				if err := s3utils.SaveErrorObject(ctx, params.S3Client, s3Config, event.S3.Bucket.Name, event.S3.Object.Key, decodedKey); err != nil {
					log.Printf("[%s] failed to move object to error location: %v", buildStamp, err)
				}
				continue
			}

			// try to get the EXIF timestamp for the object
			tstamp, err := processor.GetImgTimeStamp(imageBytes, decodedKey)
			if err != nil {
				log.Printf("[%s] failed to obtain timestamp: %v", buildStamp, err)
				if err := s3utils.SaveErrorObject(ctx, params.S3Client, s3Config, event.S3.Bucket.Name, event.S3.Object.Key, decodedKey); err != nil {
					log.Printf("[%s] failed to move object to error location: %v", buildStamp, err)
				}
				continue
			}

			// use the EXIF timestamp and the supplied key to create a destination key, then move it
			if err = s3utils.ProcessAndMoveObject(ctx, params.S3Client, s3Config, event.S3.Bucket.Name, event.S3.Object.Key, decodedKey, tstamp, params.DestinationBucket); err != nil {
				log.Printf("[%s] failed to move object: %v", buildStamp, err)
				continue
			}

			log.Printf("[%s] Processed request for : object %s/%s", buildStamp, event.S3.Bucket.Name, decodedKey)
		}
	}

	return 0, nil
}

// main function invoked when the lambda is launched
func main() {
	// v2 requires a context even for configuration loading
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(params.Region),
	)
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	params.S3Client = s3.NewFromConfig(cfg)

	log.Printf("[%s] Registering handler for photo-lambda...", buildStamp)
	lambda.Start(HandleLambdaEvent)
}
