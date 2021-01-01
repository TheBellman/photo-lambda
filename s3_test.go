package main

import (
	"testing"
	"time"
)

func Test_validatePrefix(t *testing.T) {
	type args struct {
		prefix string
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "simple", args: args{prefix: ""}, want: DefaultPrefix},
		{name: "complex", args: args{prefix: "folder"}, want: "folder/"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := validatePrefix(tt.args.prefix); got != tt.want {
				t.Errorf("extractName() = %v, want %v", got, tt.want)
			}
		})
	}
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
