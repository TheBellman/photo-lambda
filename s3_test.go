package main

import (
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/service/s3"
	"testing"
	"time"
)

type mockS3 struct{}

var jpegMime = "image/jpeg"
var txtMime = "text/plain"

func (f *mockS3) DeleteObject(input *s3.DeleteObjectInput) (*s3.DeleteObjectOutput, error) {
	if *input.Key == "case3" {
		return nil, fmt.Errorf("case3 should fail")
	}
	return &s3.DeleteObjectOutput{}, nil
}

func (f *mockS3) WaitUntilObjectExists(input *s3.HeadObjectInput) error {
	if *input.Key == "case2" {
		return fmt.Errorf("case2 should fail")
	}
	return nil
}

func (f *mockS3) CopyObject(input *s3.CopyObjectInput) (*s3.CopyObjectOutput, error) {
	if *input.Key == "case1" {
		return nil, fmt.Errorf("case1 should fail")
	}
	return &s3.CopyObjectOutput{}, nil
}

func (f *mockS3) GetObject(input *s3.GetObjectInput) (*s3.GetObjectOutput, error) {
	if *input.Key == "key/good.jpeg" {
		return &s3.GetObjectOutput{
			ContentType: &jpegMime,
			Body: testFileReader(),
		}, nil
	}

	if *input.Key == "key/bad.jpeg" {
		return &s3.GetObjectOutput{
			ContentType: &txtMime,
			Body: testFileReader(),
		}, nil
	}

	return nil, errors.New("unexpected test key provided")
}

func Test_makeNewKey(t *testing.T) {
	type args struct {
		key    string
		tstamp *time.Time
	}

	stamp := time.Date(2020, 12, 23, 16, 20, 0, 0, time.UTC)

	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "simple", args: args{key: "fred", tstamp: &stamp}, want: "photos/2020/12/23/fred"},
		{name: "complex", args: args{key: "folder/subfolder/mary", tstamp: &stamp}, want: "photos/2020/12/23/mary"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := makeNewKey(tt.args.key, tt.args.tstamp); got != tt.want {
				t.Errorf("extractName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_extractName(t *testing.T) {
	type args struct {
		key string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "empty", args: args{key: ""}, want: ""},
		{name: "simple", args: args{key: "fred"}, want: "fred"},
		{name: "complex", args: args{key: "fred/mary"}, want: "mary"},
		{name: "weird", args: args{key: "fred/mary/"}, want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := extractName(tt.args.key); got != tt.want {
				t.Errorf("extractName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getImageReader(t *testing.T) {
	mock := mockS3{}
	_, err := getImageReader(&mock, "bucket", "key/good.jpeg")
	if err != nil {
		t.Errorf("Received an unexpected error: %v", err)
	}

	_, err = getImageReader(&mock, "bucket", "key/bad.jpeg")
	if err == nil {
		t.Errorf("Did not get an error when expected")
	}
}

func Test_moveObject(t *testing.T) {
	mock := mockS3{}
	// case 1 - copy failed, expect error
	err := moveObject(&mock, "sourceBucket", "sourceKey", "destBucket", "case1")
	if err == nil {
		t.Errorf("Did not get an error for copy failure when one was expected")
	}

	// case 2 - wait failed, expect error
	err = moveObject(&mock, "sourceBucket", "sourceKey", "destBucket", "case2")
	if err == nil {
		t.Errorf("Did not get an error for wait failure when one was expected")
	}

	// case 3 - delete failed, expect error
	err = moveObject(&mock, "sourceBucket", "case3", "destBucket", "destKey")
	if err == nil {
		t.Errorf("Did not get an error for delete failure when one was expected")
	}

	// case 4 - no failures, expect no errors
	err = moveObject(&mock, "sourceBucket", "sourceKey", "destBucket", "case4")
	if err != nil {
		t.Errorf("Got an unexpected error for the no-fail case: %v", err)
	}

}