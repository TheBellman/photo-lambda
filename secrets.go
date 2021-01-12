package main

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
)

// getWasabiSecret tries to use the supplied session to retrieve the wasabi secret values
func getWasabiSecret(sess *session.Session) (key string, secret string, err error) {
	client := secretsmanager.New(sess)
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
