package main

import (
	"github.com/rwcarlsen/goexif/exif"
	"github.com/rwcarlsen/goexif/tiff"
	"io"
	"log"
	"time"
)

// getImgTimeStamp tries to get the EXIF timestamp for the image the supplied reader refers to.
// it will return an error and nil Time if the object cannot be retrieved, or is not a JPEG. If there are
// problems obtaining a meaningful timestamp from the JPEG, it will return the current time.
func getImgTimeStamp(r io.Reader) (*time.Time, error) {

	metaData, err := exif.Decode(r)
	if err != nil {
		log.Printf("Failed to get metadata from image file: %v", err)
		return nil, err
	}

	var imgDate *tiff.Tag

	imgDate, err = metaData.Get("DateTimeOriginal")
	if err != nil {
		imgDate, err = metaData.Get("DateTimeDigitized")
		if err != nil {
			imgDate, err = metaData.Get("DateTime")
		}
	}

	if imgDate != nil {
		t, err := time.Parse("\"2006:01:02 15:04:05\"", imgDate.String())
		if err != nil {
			log.Printf("Failed to parse imgDate: %v", err)
			t = time.Now()
			return &t, nil
		}
		return &t, nil
	} else {
		t := time.Now()
		return &t, nil
	}
}

