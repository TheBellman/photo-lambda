package processor

import (
	"bytes"
	"context"
	"fmt"
	_ "image/jpeg"
	"io"
	"log"
	"strings"
	"time"

	"github.com/FlavioCFOliveira/GoMetadata/format/raw/orf"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/evanoberholster/imagemeta"
	"github.com/rwcarlsen/goexif/exif"
)

// S3Service updated for v2 signatures
type S3Service interface {
	GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
	CopyObject(ctx context.Context, params *s3.CopyObjectInput, optFns ...func(*s3.Options)) (*s3.CopyObjectOutput, error)
	DeleteObject(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error)
	// ... add others if needed
}

// GetImage retrieves the byte contents of a specified reader
func GetImage(r io.Reader) (*[]byte, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return &[]byte{}, err
	}
	return &data, nil
}

func GetImageReader(ctx context.Context, service S3Service, bucket string, key string) (io.Reader, error) {
	result, err := service.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	})
	if err != nil {
		return nil, err
	}
	return result.Body, nil
}

// GetImgTimeStamp tries to get the EXIF timestamp for the image the supplied reader refers to.
// it will return an error and nil Time if the object cannot be retrieved. If there are
// problems obtaining a meaningful timestamp from the file, it will return the current time.
func GetImgTimeStamp(image *[]byte, filename string) (*time.Time, error) {

	if strings.HasSuffix(strings.ToLower(filename), ".orf") {
		return getORFDate(image)
	}

	metaData, err := imagemeta.Decode(bytes.NewReader(*image))
	if err != nil {
		log.Printf("Failed to get metadata from image file: %v", err)
		return nil, err
	}

	if !metaData.OriginalDate().IsZero() {
		t := metaData.OriginalDate()
		return &t, nil
	}

	if !metaData.DigitizedDate().IsZero() {
		t := metaData.DigitizedDate()
		return &t, nil
	}

	if !metaData.ModifyDate().IsZero() {
		t := metaData.ModifyDate()
		return &t, nil
	}

	t := time.Now()
	return &t, nil
}

// getORFDate extracts the EXIF timestamp from an Olympus RAW (.orf) image.
// It uses a specialized method to handle the unique EXIF structure of ORF files,
// specifically addressing the "IIRO" magic byte patch via the GoMetadata library.
// If no valid timestamp is found, it returns the current time.
func getORFDate(image *[]byte) (*time.Time, error) {
	reader := bytes.NewReader(*image)

	// Extract raw EXIF payload (handles the "IIRO" magic byte patch)
	rawExif, _, _, err := orf.Extract(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to extract metadata: %w", err)
	}

	if rawExif == nil {
		t := time.Now()
		return &t, nil
	}

	// Decode the raw bytes into tags
	exifData, err := exif.Decode(bytes.NewReader(rawExif))
	if err != nil {
		return nil, fmt.Errorf("failed to decode exif: %w", err)
	}

	// Retrieve the DateTimeOriginal, DateTimeDigitized or DateTime tag depdending which is available
	tm, err := exifData.DateTime()
	if err != nil {
		t := time.Now()
		return &t, nil
	}

	return &tm, nil
}
