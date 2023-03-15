package image

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

const defaultImageFiletype = ".png"

var stdout io.Writer = os.Stdout

type Client struct {
	httpClient      *http.Client
	defaultFilename string
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Transport: &http.Transport{
				DialContext: (&net.Dialer{
					Timeout:   30 * time.Second,
					KeepAlive: 30 * time.Second,
				}).DialContext,
				TLSHandshakeTimeout:   10 * time.Second,
				ResponseHeaderTimeout: 10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			},
		},
	}
}

func (c *Client) SetDefaultFilename(defaultFilename string) {
	(*c).defaultFilename = defaultFilename
}

func (c *Client) DownloadImage(ctx context.Context, url string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	// Get image data
	response, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		return fmt.Errorf("error: HTTP status code %d for url %s received", response.StatusCode, url)
	}

	// Get unique filename
	var filename string
	filename, err = getFilename((*c).defaultFilename, url, response)
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
	// TODO: This does not work for openai urls, maybe we can find a different way?
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
