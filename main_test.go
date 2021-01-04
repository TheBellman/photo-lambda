package main

import (
	"testing"
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
