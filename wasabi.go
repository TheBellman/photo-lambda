package main

import (
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