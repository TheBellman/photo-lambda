package main

import (
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/service/s3"
	"testing"
	"time"
)

func Test_validateRegion(t *testing.T) {
	type args struct {
		region string
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "empty", args: args{region: ""}, want: DefaultRegion},
		{name: "nonempty", args: args{region: "us-east-1"}, want: "us-east-1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := validateRegion(tt.args.region); got != tt.want {
				t.Errorf("extractName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_validatePrefix(t *testing.T) {
	type args struct {
		prefix        string
		defaultPrefix string
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "empty", args: args{prefix: "", defaultPrefix: DefaultSrcPrefix}, want: DefaultSrcPrefix},
		{name: "nonempty", args: args{prefix: "folder", defaultPrefix: DefaultSrcPrefix}, want: "folder/"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := validatePrefix(tt.args.prefix, DefaultSrcPrefix); got != tt.want {
				t.Errorf("extractName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_validateDestination(t *testing.T) {
	type args struct {
		dest string
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "empty", args: args{dest: ""}, want: DefaultBucket},
		{name: "nnonempty", args: args{dest: "mybucket"}, want: "mybucket"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := validateDestination(tt.args.dest); got != tt.want {
				t.Errorf("extractName() = %v, want %v", got, tt.want)
			}
		})
	}
}


type mockS3 struct{}

var jpegMime = "image/jpeg"
var txtMime = "text/plain"

func (f *mockS3) PutObject(input *s3.PutObjectInput) (*s3.PutObjectOutput, error) {
	return &s3.PutObjectOutput{}, nil
}

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
			Body:        testFileReader("./test.jpeg"),
		}, nil
	}

	if *input.Key == "key/bad.jpeg" {
		return &s3.GetObjectOutput{
			ContentType: &txtMime,
			Body:        testFileReader("./test.jpeg"),
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
		{name: "simple", args: args{key: "import/fred", tstamp: &stamp}, want: "photos/2020/12/23/fred"},
		{name: "robert", args: args{key: "import/robert/img2.jpg", tstamp: &stamp}, want: "photos/robert/2020/12/23/img2.jpg"},
		{name: "delia", args: args{key: "import/delia/img1.jpg", tstamp: &stamp}, want: "photos/delia/2020/12/23/img1.jpg"},
		{name: "complex", args: args{key: "import/folder/subfolder/mary", tstamp: &stamp}, want: "photos/folder/subfolder/2020/12/23/mary"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := makeNewKey(tt.args.key, tt.args.tstamp); got != tt.want {
				t.Errorf("extractName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_makeErrKey(t *testing.T) {
	type args struct {
		key    string
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "simple", args: args{key: "import/fred"}, want: "errors/fred"},
		{name: "robert", args: args{key: "import/robert/img2.jpg"}, want: "errors/robert/img2.jpg"},
		{name: "delia", args: args{key: "import/delia/img1.jpg"}, want: "errors/delia/img1.jpg"},
		{name: "complex", args: args{key: "import/folder/subfolder/mary"}, want: "errors/folder/subfolder/mary"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := makeErrKey(tt.args.key); got != tt.want {
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
