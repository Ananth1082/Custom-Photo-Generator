package utils

import (
	"errors"
	"image"
	"log"
	"net/http"
	"regexp"
)

// GetImage retrieves and decodes an image from a given Google Drive sharing link.
func GetImage(url string) (image.Image, error) {
	// Improved regex pattern to match the file ID from a variety of Google Drive link formats
	var idregex = regexp.MustCompile(`(?:file/d/|open\?id=|uc\?export=download&id=|d/|id=)([^/&?#]+)`)

	// Get the download link using the file ID
	durl, err := getDownloadLink(url, idregex)
	if err != nil {
		log.Printf("Failed to get download link: %v", err)
		return nil, err
	}

	// Perform HTTP GET request to fetch the image
	res, err := http.Get(durl)
	if err != nil {
		log.Printf("HTTP request failed: %v", err)
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		err := errors.New("failed to fetch image, status code: " + res.Status)
		log.Printf("Non-OK HTTP status: %v", err)
		return nil, err
	}

	// Decode the image from the response body
	img, _, err := image.Decode(res.Body)
	if err != nil {
		log.Printf("Failed to decode image: %v", err)
		return nil, err
	}
	return img, nil
}

// getDownloadLink converts a Google Drive sharing link into a direct download link.
func getDownloadLink(url string, rgx *regexp.Regexp) (string, error) {
	// Extract the file ID using regex
	matches := rgx.FindStringSubmatch(url)
	if len(matches) < 2 {
		err := errors.New("invalid Google Drive sharing link")
		log.Printf("Regex match failed: %v", err)
		return "", err
	}
	fileID := matches[1]

	// Construct the direct download link
	return "https://drive.google.com/uc?export=download&id=" + fileID, nil
}
