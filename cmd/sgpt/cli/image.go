package cli

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"flag"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/peterbourgon/ff/v3/ffcli"
	"github.com/sashabaranov/go-openai"
	"github.com/tbckr/sgpt"
)

const defaultImageFiletype = ".png"

var imageCmd = &ffcli.Command{
	Name:       "image",
	ShortUsage: "sgpt image [command flags] <prompt>",
	ShortHelp:  "Create an AI generated image with dalle.",
	LongHelp: strings.TrimSpace(`
Create an AI generated image with the DALLE API. 
`),
	Exec: runImage,
	FlagSet: (func() *flag.FlagSet {
		fs := newFlagSet("image")
		fs.IntVar(&imageArgs.count, "count", 1, "Number of images to generate")
		fs.StringVar(&imageArgs.size, "size", openai.CreateImageSize256x256, "Image size")
		fs.BoolVar(&imageArgs.download, "download", false, "Download generated images")
		fs.StringVar(&imageArgs.outputFilename, "output", "", "Filename including path to file - might be used for base name, if multiple images are created")
		return fs
	})(),
}

var imageArgs struct {
	count          int
	size           string
	download       bool
	outputFilename string
}

func runImage(ctx context.Context, args []string) error {
	// Check, if prompt was provided via command line
	if len(args) != 1 {
		return ErrMissingPrompt
	}
	prompt := args[0]

	options := sgpt.ImageOptions{
		Count: imageArgs.count,
	}
	if err := sgpt.ValidateImageOptions(options); err != nil {
		return err
	}

	var responseFormat string
	if imageArgs.download {
		responseFormat = openai.CreateImageResponseFormatB64JSON
	} else {
		responseFormat = openai.CreateImageResponseFormatURL
	}

	client, err := sgpt.CreateClient()
	if err != nil {
		return err
	}

	var imageData []string
	imageData, err = sgpt.GetImage(ctx, client, options, prompt, responseFormat)
	if err != nil {
		return err
	}
	return handleImageURLs(imageData)
}

func handleImageURLs(images []string) error {
	for _, data := range images {
		if imageArgs.download {
			// Save base64 encoded data to file
			if err := save2File(data); err != nil {
				return err
			}
		} else {
			// Print url to stdout
			if _, err := fmt.Fprintln(stdout, data); err != nil {
				return err
			}
		}
	}
	return nil
}

func save2File(imageData string) error {
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
	filename, err = getFilename()
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

	_, err = fmt.Fprintln(stdout, filename)
	return err
}

func getFilename() (string, error) {
	var filename string
	var err error

	// Generate a filename, if no default filename was provided
	if imageArgs.outputFilename == "" {
		filename, err = generateFilename()
		if err != nil {
			return "", err
		}
	} else {
		filename = imageArgs.outputFilename
	}

	// Check, if file already exists
	var exists bool
	baseFilename := strings.Clone(filename)
	for {
		exists, err = fileExists(filename)
		if err != nil {
			return "", err
		}
		// If file exists, append random string
		if exists {
			filename, err = makeUnique(baseFilename)
			if err != nil {
				return "", err
			}
		} else {
			break
		}
	}
	return filename, nil
}

func generateFilename() (string, error) {
	u, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}
	return u.String() + defaultImageFiletype, nil
}

func fileExists(filename string) (bool, error) {
	// If file exists, append random string
	if _, err := os.Stat(filename); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func makeUnique(filename string) (string, error) {
	// Get filename and extension
	ext := filepath.Ext(filename)
	name := filename[:len(filename)-len(ext)]

	if ext == "" {
		ext = defaultImageFiletype
	}

	// Generate a random byte slice
	b := make([]byte, 6)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	// Encode the byte slice as a string using base64 encoding
	randomSuffix := base64.URLEncoding.EncodeToString(b)

	if name == "" {
		return fmt.Sprintf("%s%s", randomSuffix, ext), nil
	}
	return fmt.Sprintf("%s_%s%s", name, randomSuffix, ext), nil
}
