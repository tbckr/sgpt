package file

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/png"
	"io"
	"os"
)

func SaveEncodedImage(defaultFilename, imageData string, out io.Writer) error {
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

	var filename string
	filename, err = getFilename(defaultFilename)
	if err != nil {
		return err
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

	_, err = fmt.Fprintln(out, filename)
	return err
}
