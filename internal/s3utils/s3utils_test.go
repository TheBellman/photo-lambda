package s3utils

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/TheBellman/photo-lambda/internal/testutils"
)

// getTestDataReader helps find the test files relative to this test file
func getTestDataReader(name string) (io.ReadCloser, error) {
	// Relative path from internal/s3utils/ to testdata/
	path := filepath.Join("..", "..", "testdata", name)
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func Test_makeNewKey(t *testing.T) {
	config := Config{
		SourcePrefix:      "import/",
		DestinationPrefix: "photos/",
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
			if got := makeNewKey(config, tt.args.key, tt.args.tstamp); got != tt.want {
				t.Errorf("makeNewKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_makeErrKey(t *testing.T) {
	config := Config{
		SourcePrefix: "import/",
		ErrorPrefix:  "errors/",
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
			if got := makeErrKey(config, tt.args.key); got != tt.want {
				t.Errorf("makeErrKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_ProcessAndMoveObject(t *testing.T) {
	mock := &testutils.MockS3{TestDataReader: getTestDataReader}
	ctx := context.TODO()
	config := Config{
		SourcePrefix:      "import/",
		DestinationPrefix: "photos/",
	}
	stamp := time.Date(2020, 12, 23, 16, 20, 0, 0, time.UTC)

	// case 1 - copy failed, expect error
	err := ProcessAndMoveObject(ctx, mock, config, "sourceBucket", "sourceKey", "import/case1", &stamp, "destBucket")
	if err == nil {
		t.Errorf("Did not get an error for copy failure when one was expected")
	}

	// case 3 - delete failed, expect error
	err = ProcessAndMoveObject(ctx, mock, config, "sourceBucket", "case3", "import/destKey", &stamp, "destBucket")
	if err == nil {
		t.Errorf("Did not get an error for delete failure when one was expected")
	}

	// case 4 - no failures, expect no errors
	err = ProcessAndMoveObject(ctx, mock, config, "sourceBucket", "sourceKey", "import/case4", &stamp, "destBucket")
	if err != nil {
		t.Errorf("Got an unexpected error for the no-fail case: %v", err)
	}
}
