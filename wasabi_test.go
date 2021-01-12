package main

import (
	"testing"
)

func Test_makeWasabiConfig(t *testing.T) {
	params := &runtimeParameters{
		WasabiKey:         "theKey",
		WasabiSecret:      "theSecret",
		WasabiRegion:      "eu-central-1",
		WasabiBucket:      "theBucket",
	}

	result := makeWasabiConfig(params)
	if result == nil {
		t.Error("makeWasabiConfig() gave a nil config when it should be impossible")
	}

	if *result.Region != "eu-central-1" {
		t.Errorf("makeWasabiConfig() incorrect region %q set", *result.Region)
	}

	if *result.Endpoint != "https://s3.eu-central-1.wasabisys.com" {
		t.Errorf("makeWasabiConfig() incorrect endpoint %q set", *result.Endpoint)
	}

	if result.Credentials == nil  {
		t.Error("makeWasabiConfig() did not set credentials")
	}
}

func Test_makeWasabiClient(t *testing.T) {
	params := &runtimeParameters{
		WasabiKey:         "theKey",
		WasabiSecret:      "theSecret",
		WasabiRegion:      "eu-central-1",
		WasabiBucket:      "theBucket",
	}

	result, err := makeWasabiClient(params)
	if err != nil {
		t.Errorf("makeWasabiClient() unexpected error %v", err)
	}
	if result == nil {
		t.Error("makeWasabiClient() gave a nil config when it should be impossible")
	}
}