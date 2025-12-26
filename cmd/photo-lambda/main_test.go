package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/TheBellman/photo-lambda/internal/testutils"
)

// getTestDataReader helps find the test files relative to this test file
func getTestDataReader(name string) io.ReadCloser {
	// Relative path from cmd/photo-lambda/ to testdata/
	path := filepath.Join("..", "..", "testdata", name)
	f, err := os.Open(path)
	if err != nil {
		// Using panic in test setup is okay if the files are mandatory
		panic(fmt.Sprintf("failed to open test file %s: %v", path, err))
	}
	return f
}

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

func Test_moveObject(t *testing.T) {
	mock := &testutils.MockS3{TestDataReader: getTestDataReader}

	// case 1 - copy failed, expect error
	err := moveObject(mock, "sourceBucket", "sourceKey", "destBucket", "case1")
	if err == nil {
		t.Errorf("Did not get an error for copy failure when one was expected")
	}

	// case 2 - wait failed, expect error
	err = moveObject(mock, "sourceBucket", "sourceKey", "destBucket", "case2")
	if err == nil {
		t.Errorf("Did not get an error for wait failure when one was expected")
	}

	// case 3 - delete failed, expect error
	err = moveObject(mock, "sourceBucket", "case3", "destBucket", "destKey")
	if err == nil {
		t.Errorf("Did not get an error for delete failure when one was expected")
	}

	// case 4 - no failures, expect no errors
	err = moveObject(mock, "sourceBucket", "sourceKey", "destBucket", "case4")
	if err != nil {
		t.Errorf("Got an unexpected error for the no-fail case: %v", err)
	}
}

func Test_makeNewKey(t *testing.T) {
	// Ensure params are initialized for the test
	params = &runtimeParameters{
		SourcePrefix:      DefaultSrcPrefix,
		DestinationPrefix: DefaultDestPrefix,
	}
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
	// Initialize global params for the test environment
	params = &runtimeParameters{
		SourcePrefix: DefaultSrcPrefix, // "import/"
		ErrorPrefix:  DefaultErrPrefix, // "errors/"
	}
	
	type args struct {
		key string
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
				t.Errorf("makeErrKey() = %v, want %v", got, tt.want) // Change 'extractName' to 'makeErrKey'
			}
		})
	}
}
