package image

import (
	"bytes"
	"encoding/base64"
	"errors"
	"image"
	"image/png"
	"os"

	"github.com/tbckr/sgpt/filesystem"
)

const DefaultExtension = ".png"

func SaveB64EncodedImage(filename, imageData string) error {
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
	exists, err = filesystem.FileExists(filename)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("file already exists")
	}

	var file *os.File
	file, err = os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	if err = png.Encode(file, img); err != nil {
		return err
	}
	return nil
}
