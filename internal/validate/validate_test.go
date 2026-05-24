package validate

import (
	"testing"
)

func TestRegion(t *testing.T) {
	defaultRegion := "eu-west-2"
	type args struct {
		region string
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "empty", args: args{region: ""}, want: defaultRegion},
		{name: "nonempty", args: args{region: "us-east-1"}, want: "us-east-1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Region(tt.args.region, defaultRegion); got != tt.want {
				t.Errorf("Region() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPrefix(t *testing.T) {
	defaultPrefix := "import/"
	type args struct {
		prefix        string
		defaultPrefix string
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "empty", args: args{prefix: "", defaultPrefix: defaultPrefix}, want: defaultPrefix},
		{name: "nonempty", args: args{prefix: "folder", defaultPrefix: defaultPrefix}, want: "folder/"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Prefix(tt.args.prefix, defaultPrefix); got != tt.want {
				t.Errorf("Prefix() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDestination(t *testing.T) {
	defaultBucket := "NOSUCHBUCKET"
	type args struct {
		dest string
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "empty", args: args{dest: ""}, want: defaultBucket},
		{name: "nonempty", args: args{dest: "mybucket"}, want: "mybucket"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Destination(tt.args.dest, defaultBucket); got != tt.want {
				t.Errorf("Destination() = %v, want %v", got, tt.want)
			}
		})
	}
}
