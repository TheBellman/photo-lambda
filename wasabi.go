package main

import (
	"bytes"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	session2 "github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"log"
)

// makeWasabiConfig creates a suitable s3 client configuration from the parameters
func makeWasabiConfig(params *runtimeParameters) *aws.Config {
	s3Config := &aws.Config{
		Credentials:                       credentials.NewStaticCredentials(params.WasabiKey, params.WasabiSecret, ""),
		Endpoint:                          aws.String(fmt.Sprintf("https://s3.%s.wasabisys.com", params.WasabiRegion)),
		Region:                            aws.String(params.WasabiRegion),
		S3ForcePathStyle:                  aws.Bool(true),
	}

	return s3Config
}

// makeWasabiClient creates an s3 client from the parameters
func makeWasabiClient(params *runtimeParameters) (*s3.S3, error){
	config := makeWasabiConfig(params)
	session, err := session2.NewSession(config)
	if err!=nil {
		return nil, fmt.Errorf("failed to create a session for Wasabi: %v", err)
	}

	client := s3.New(session)
	return client, nil
}

// makeLazyWasabiClient does a lazy instantiation of the wasabi client and adds it to the runtime parameter set
func makeLazyWasabiClient(params *runtimeParameters) {
	// try to do a lazy fetch of the wasabi secret.
	lazyGetSecret(params)

	if params.WasabiService == nil {
		client, err := makeWasabiClient(params)
		if err != nil {
			log.Fatalf("failed to create a client for Wasabi: %v", err)
		}
		params.WasabiService = client
	}
}

// saveToWasabi uses the current configuration to write the supplied image bytes to the target key
func saveToWasabi(params *runtimeParameters, image *[]byte, key string) error {
	if params.WasabiService == nil {
		return fmt.Errorf("no service has been provided to write to wasabi")
	}

	result, err := params.WasabiService.PutObject(&s3.PutObjectInput{
		Body:                      bytes.NewReader(*image),
		Bucket:                    aws.String(params.WasabiBucket),
		ContentType:               aws.String(JPEG),
		Key:                       aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("writing to wasabi failed: %v", err)
	}
	log.Printf("copied to wasabi %q with etag %q", key, *result.ETag)

	return nil
}