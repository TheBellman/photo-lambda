package main

import (
	"bytes"
	"fmt"
	"github.com/disintegration/imaging"
	"image/jpeg"
	"io/ioutil"

	//"github.com/disintegration/imaging"
	"github.com/rwcarlsen/goexif/exif"
	"github.com/rwcarlsen/goexif/tiff"
	"image"
	_ "image/jpeg"
	"io"
	"log"
	"time"
)

// getImage retrieves the byte contents of a specified reader
func getImage(r io.Reader) (*[]byte, error) {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return &[]byte{}, err
	}
	return &data, nil
}

// resizeImage attempts to resize the supplied image (assuming the bytes represent a
// jpeg) and hand back a new byte array representing the smaller jpeg
func resizeImage(origImg *[]byte) (*[]byte, error) {
	imgConf, _, err := image.DecodeConfig(bytes.NewReader(*origImg))
	if err != nil {
		return &[]byte{}, fmt.Errorf("failed to decode byte stream as a jpeg: %v", err)
	}

	original, _, err := image.Decode(bytes.NewReader(*origImg))
	if err != nil {
		return &[]byte{}, fmt.Errorf("failed to decode byte stream as a jpeg: %v", err)
	}

	// a new width/height of zero means "retain aspect ratio", so we only set one
	newWidth := 0
	newHeight := 0
	if imgConf.Width > imgConf.Height {
		newWidth = ThumbnailSize
	} else {
		newHeight = ThumbnailSize
	}

	newImage := imaging.Resize(original, newWidth, newHeight, imaging.Lanczos)

	buf := new(bytes.Buffer)
	err = jpeg.Encode(buf, newImage, nil)
	if err != nil {
		return &[]byte{}, fmt.Errorf("failed to encode resized image as jpeg: %v", err)
	}

	result := buf.Bytes()
	return &result, nil
}

// getImgTimeStamp tries to get the EXIF timestamp for the image the supplied reader refers to.
// it will return an error and nil Time if the object cannot be retrieved, or is not a JPEG. If there are
// problems obtaining a meaningful timestamp from the JPEG, it will return the current time.
func getImgTimeStamp(image *[]byte) (*time.Time, error) {

	metaData, err := exif.Decode(bytes.NewReader(*image))
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
