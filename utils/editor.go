package utils

import (
	models "CustomPhotoGenerator/m-v0/Models"
	"image"
	"image/draw"
	"log"

	scaleDraw "golang.org/x/image/draw"

	"github.com/golang/freetype"
)

// Add a label to an image at the specified location
func AddLabel(img image.Image, td models.TextDetails, location image.Point, content string) (image.Image, error) {
	canvas := image.NewRGBA(img.Bounds())
	draw.Draw(canvas, canvas.Bounds(), img, image.Point{}, draw.Src)

	font, err := LoadFont(td)
	if err != nil {
		return nil, err
	}

	c := freetype.NewContext()
	c.SetDPI(72)
	c.SetFont(font)
	c.SetFontSize(float64(td.Size))
	c.SetClip(canvas.Bounds())
	c.SetDst(canvas)
	c.SetSrc(&image.Uniform{td.Color})

	pt := freetype.Pt(location.X, location.Y+int(c.PointToFixed(float64(td.Size))>>6))
	_, err = c.DrawString(content, pt)
	if err != nil {
		return nil, err
	}
	return canvas, nil
}
func AddImage(baseImg image.Image, img image.Image, location image.Point) image.Image {
	// Ensure both images are not nil
	if baseImg == nil || img == nil {
		log.Println("One of the images is nil")
		return baseImg
	}

	// Convert baseImg to *image.RGBA to draw on it
	baseImgRGBA, ok := baseImg.(*image.RGBA)
	if !ok {
		// If baseImg is not *image.RGBA, create a new *image.RGBA from it
		b := baseImg.Bounds()
		newBaseImg := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
		draw.Draw(newBaseImg, newBaseImg.Bounds(), baseImg, b.Min, draw.Src)
		baseImgRGBA = newBaseImg
	}

	// Draw img onto baseImg at the specified location
	draw.Draw(baseImgRGBA, img.Bounds().Add(location), img, image.Point{}, draw.Over)

	return baseImgRGBA
}

func FitImage(img image.Image, newHeight, newWidth int) image.Image {
	// Create a new blank image with the new dimensions
	dstImage := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
	// Resize the image using BiLinear interpolation
	scaleDraw.BiLinear.Scale(dstImage, dstImage.Bounds(), img, img.Bounds(), draw.Over, nil)
	return dstImage
}
