package utils

import (
	"bytes"
	"image"
	"image/png"
)

// Save image to the provided filename
func ImageToBytes(img image.Image) ([]byte, error) {
	var buf bytes.Buffer
	err := png.Encode(&buf, img)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
