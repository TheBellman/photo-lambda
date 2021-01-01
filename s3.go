package main

import (
	"path/filepath"
	"strings"
	"time"
)

// validatePrefix coerces the environmental variable into a usable prefix, by adding a "/" if necessary or setting it to
// the default prefix. It returns the coerced prefix
func validatePrefix(photoPrefix string) string {
	if !strings.HasSuffix(photoPrefix, "/") {
		if photoPrefix == "" {
			photoPrefix = DefaultPrefix
		} else {
			photoPrefix += "/"
		}
	}
	return photoPrefix
}

// extractName gets the last part of the S3 key
func extractName(key string) string {
	if key == "" || strings.HasSuffix(key, "/") {
		return ""
	}
	return filepath.Base(key)
}

// makeNewKey will assemble the target key for a provided incoming object key, and the timestamp
func makeNewKey(key string, tstamp *time.Time) string {
	return params.DestinationPrefix + tstamp.Format("2006/01/02/") + extractName(key)
}
