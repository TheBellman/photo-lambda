package main

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"log"
)

// secretService helps with mocking access to secrets manager
type secretService interface {
	GetSecretValue(input *secretsmanager.GetSecretValueInput) (*secretsmanager.GetSecretValueOutput, error)
}

// getWasabiSecret tries to use the supplied session to retrieve the wasabi secret values
func getWasabiSecret(client secretService) (key string, secret string, err error) {
	result, err := client.GetSecretValue(&secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(WasabiSecret),
	})
	if err != nil {
		return "", "", fmt.Errorf("failed to read the wasabi secret  %v", err)
	}

	var values map[string]string
	err = json.Unmarshal([]byte(*result.SecretString), &values)
	if err != nil {
		return "", "", fmt.Errorf("failed to read the secret JSON %v", err)
	}

	return values["ACCESS_KEY_ID"], values["SECRET_ACCESS_KEY"], nil
}

// lazyGetSecret tries to fetch the wasabi secret values only if they are not already in the parameters
func lazyGetSecret() {
	if params.WasabiSecret == "" {
		key, secret, err := getWasabiSecret(secretsmanager.New(params.Session))
		if err != nil {
			log.Fatal("Failed to fetch Wasabi secrets", err)
		}
		params.WasabiKey = key
		params.WasabiSecret = secret
		log.Printf("Successfully fetched Wasabi secret [%s]", key)
	}
}
