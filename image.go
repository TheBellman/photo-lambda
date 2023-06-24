package main

import (
	"bytes"
	"github.com/evanoberholster/imagemeta"
	_ "image/jpeg"
	"io"
	"log"
	"time"
)

// getImage retrieves the byte contents of a specified reader
func getImage(r io.Reader) (*[]byte, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return &[]byte{}, err
	}
	return &data, nil
}

// getImgTimeStamp tries to get the EXIF timestamp for the image the supplied reader refers to.
// it will return an error and nil Time if the object cannot be retrieved. If there are
// problems obtaining a meaningful timestamp from the file, it will return the current time.
func getImgTimeStamp(image *[]byte) (*time.Time, error) {

	metaData, err := imagemeta.Decode(bytes.NewReader(*image))
	if err != nil {
		log.Printf("Failed to get metadata from image file: %v", err)
		return nil, err
	}

	if !metaData.DateTimeOriginal().IsZero() {
		t := metaData.DateTimeOriginal()
		return &t, nil
	}

	if !metaData.CreateDate().IsZero() {
		t := metaData.CreateDate()
		return &t, nil
	}

	if !metaData.ModifyDate().IsZero() {
		t := metaData.ModifyDate()
		return &t, nil
	}

	t := time.Now()
	return &t, nil
}
