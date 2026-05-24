package validate

import (
	"strings"
)

// validateRegion will provide the default region if no region is set
func Region(region string, defaultRegion string) string {
	if region == "" {
		return defaultRegion
	} else {
		return region
	}
}

// Prefix coerces the environmental variable into a usable prefix, by adding a "/" if necessary or setting it to
// the default prefix. It returns the coerced prefix
func Prefix(photoPrefix string, defaultPrefix string) string {
	if !strings.HasSuffix(photoPrefix, "/") {
		if photoPrefix == "" {
			photoPrefix = defaultPrefix
		} else {
			photoPrefix += "/"
		}
	}
	return photoPrefix
}

// Destination will ensure a non-blank destination bucket
func Destination(bucket string, defaultBucket string) string {
	if bucket == "" {
		return defaultBucket
	} else {
		return bucket
	}
}
