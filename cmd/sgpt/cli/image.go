package cli

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/peterbourgon/ff/v3/ffcli"
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
		fs.IntVar(&imageArgs.count, "n", 1, "Number of images to generate")
		fs.BoolVar(&imageArgs.download, "download", false, "Download generated images")
		fs.StringVar(&imageArgs.outputFilename, "output", "", "Filename including path to file - might be used for base name, if multiple images are created")
		return fs
	})(),
}

var imageArgs struct {
	count          int
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

	client, err := sgpt.CreateClient()
	if err != nil {
		return err
	}

	var imageUrls []string
	imageUrls, err = sgpt.GetImage(ctx, client, options, prompt)
	if err != nil {
		return err
	}
	return handleImageURLs(imageUrls)
}

func handleImageURLs(imageUrls []string) error {
	if imageArgs.download {
		return downloadImages(imageUrls)
	}
	return printURLs(imageUrls)
}

func printURLs(urls []string) error {
	for _, element := range urls {
		if _, err := fmt.Fprintln(stdout, element); err != nil {
			return err
		}
	}
	return nil
}

func downloadImages(urls []string) error {
	var err error
	for _, url := range urls {
		if err = downloadImage(url); err != nil {
			return err
		}
	}
	return nil
}

func downloadImage(url string) error {
	// Get image data
	/* #nosec */
	response, err := http.Get(url)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		return fmt.Errorf("error: HTTP status code %d for url %s received", response.StatusCode, url)
	}

	// Get unique filename
	var filename string
	filename, err = getFilename(imageArgs.outputFilename, url, response)
	if err != nil {
		return err
	}

	// Print filename
	if _, err = fmt.Fprintln(stdout, filename); err != nil {
		return err
	}

	// Create file
	var file *os.File
	file, err = os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write file content
	_, err = io.Copy(file, response.Body)
	return err
}

func getFilename(defaultFilename, url string, resp *http.Response) (string, error) {
	// Generate a filename, if no default filename was provided
	var filename string
	if defaultFilename == "" {
		filename = deriveFilename(url, resp)
	} else {
		filename = defaultFilename
	}
	// Check, if file already exists
	baseFilename := strings.Clone(filename)
	for {
		exists, err := fileExists(filename)
		if err != nil {
			return "", err
		}
		// If file exists, append random string
		if exists {
			filename, err = createUniqueFilename(baseFilename)
			if err != nil {
				return "", err
			}
		} else {
			break
		}
	}
	return filename, nil
}

func deriveFilename(url string, resp *http.Response) string {
	// Check the "Content-Disposition" header for a filename
	var filename string
	if cd := resp.Header.Get("Content-Disposition"); cd != "" {
		if i := strings.Index(cd, "filename="); i >= 0 {
			filename = cd[i+len("filename="):]
			if j := strings.Index(filename, ";"); j >= 0 {
				filename = filename[:j]
			}
			filename = strings.Trim(filename, "\"")
		}
	}
	// If no filename was specified in the header, use the URL path
	// This does not work for openai urls
	//if filename == "" {
	//	filename = path.Base(url)
	//}
	// If the filename is still empty, generate a random filename
	if filename == "" {
		u, err := uuid.NewRandom()
		if err != nil {
			panic(err)
		}
		filename = u.String() + defaultImageFiletype
	}
	return filename
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

func createUniqueFilename(filename string) (string, error) {
	// Get filename and extension
	ext := filepath.Ext(filename)
	name := filename[:len(filename)-len(ext)]

	// Generate a random byte slice
	b := make([]byte, 6)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	// Encode the byte slice as a string using base64 encoding
	s := base64.URLEncoding.EncodeToString(b)

	return fmt.Sprintf("%s_%s%s", name, s, ext), nil
}
