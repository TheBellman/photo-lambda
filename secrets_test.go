package main

import (
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"testing"
)

type mockClient struct{}

var secret = "{\"ACCESS_KEY_ID\": \"key\", \"SECRET_ACCESS_KEY\": \"secret\"}"

func (f *mockClient) GetSecretValue(input *secretsmanager.GetSecretValueInput) (*secretsmanager.GetSecretValueOutput, error) {
	return &secretsmanager.GetSecretValueOutput{
		SecretString: &secret,
	}, nil
}

func Test_getWasabiSecret(t *testing.T) {
	mock := mockClient{}
	key, secret, err := getWasabiSecret(&mock)
	if err != nil {
		t.Errorf("getWasabiSecret() : %v", err)
	}

	if key != "key" || secret != "secret" {
		t.Errorf("wanted %q, %q, got %q, %q", "key", "secret", key, secret)
	}

}
