package image

import (
	"bytes"
	"encoding/base64"
	"errors"
	"image"
	"image/png"
	"os"

	"github.com/tbckr/sgpt/files"
)

const DefaultExtension = ".png"

func SaveEncodedImage(filename, imageData string) error {
	imgBytes, err := base64.StdEncoding.DecodeString(imageData)
	if err != nil {
		return err
	}

	reader := bytes.NewReader(imgBytes)
	var img image.Image
	img, err = png.Decode(reader)
	if err != nil {
		return err
	}

	var exists bool
	exists, err = files.Exists(filename)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("file already exists")
	}

	var f *os.File
	f, err = os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	if err = png.Encode(f, img); err != nil {
		return err
	}
	return nil
}
