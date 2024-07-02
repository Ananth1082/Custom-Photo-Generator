package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"

	"os"

	"sync"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	_ "golang.org/x/image/font"
)

type ishareRequest struct {
	MergedImage    string
	ConstTextBoxes []ConstTextBox
	VarTextBoxes   []VarTextBox
	VarImageBoxes  []VarImageBox
}

type TextDetails struct {
	Name     string
	IsItalic bool
	IsBold   bool
	Color    color.RGBA
	Size     int
}

type ConstTextBox struct {
	MetaDetails  TextDetails
	ConstContent string
	Location     image.Point
}
type ImageDetails struct {
	Scale struct {
		XScale float32
		YScale float32
	}
}
type VarImageBox struct {
	MetaDetails ImageDetails
	Image       []image.Image
	Location    image.Point
}

type VarTextBox struct {
	MetaDetails TextDetails
	VarContent  []string
	Location    image.Point
}

// Load image from the provided filename
func loadImage(filename string) (image.Image, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, err := png.Decode(file)
	if err != nil {
		return nil, err
	}

	return img, nil
}

// Save image to the provided filename
func saveImage(img image.Image) ([]byte, error) {
	var buf bytes.Buffer
	err := png.Encode(&buf, img)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Load font based on text details, with caching
var fontCache = sync.Map{}

func loadFont(td TextDetails) (*truetype.Font, error) {
	fontSpecifier := fmt.Sprintf("%s-%s%s",
		td.Name,
		func() string {
			if td.IsBold {
				return "Bold"
			}
			return "Regular"
		}(),
		func() string {
			if td.IsItalic {
				return "Italic"
			}
			return ""
		}(),
	)

	if font, found := fontCache.Load(fontSpecifier); found {
		return font.(*truetype.Font), nil
	}

	fontBytes, err := os.ReadFile(fmt.Sprintf("./Fonts/%s/%s.ttf", td.Name, fontSpecifier))
	if err != nil {
		return nil, fmt.Errorf("failed to load font: %v", err)
	}
	font, err := freetype.ParseFont(fontBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse font: %v", err)
	}

	fontCache.Store(fontSpecifier, font)
	return font, nil
}

// Add a label to an image at the specified location
func addLabel(img image.Image, td TextDetails, location image.Point, content string) (image.Image, error) {
	canvas := image.NewRGBA(img.Bounds())
	draw.Draw(canvas, canvas.Bounds(), img, image.Point{}, draw.Src)

	font, err := loadFont(td)
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

// Print constant content on the image
// func printConstContent(img image.Image, textData ishareRequest) (image.Image, error) {
// 	if len(textData.ConstTextBoxes) == 0 {
// 		return img, nil
// 	}

// 	for _, tb := range textData.ConstTextBoxes {
// 		var err error
// 		img, err = addLabel(img, tb.MetaDetails, tb.Location, tb.ConstContent)
// 		if err != nil {
// 			return nil, err
// 		}
// 	}
// 	err := saveImage(img, )
// 	if err != nil {
// 		return nil, err
// 	}

// 	log.Println("Constant image saved successfully.")
// 	return img, nil
// }

// Print variable content on the image
// func printVarContent(img image.Image, textData ishareRequest) ([][]byte, error) {
// 	var imagesBytes [][]byte // To store the bytes of all generated images

// 	if len(textData.VarTextBoxes) == 0 {
// 		imgBytes, err := saveImage(img)
// 		if err != nil {
// 			return nil, err
// 		}
// 		imagesBytes = append(imagesBytes, imgBytes)
// 		return imagesBytes, nil
// 	}

// 	var wg sync.WaitGroup
// 	var mu sync.Mutex                                                    // To safely append to imagesBytes
// 	errors := make(chan error, len(textData.VarTextBoxes[0].VarContent)) // For error handling

// 	for i := 0; i < len(textData.VarTextBoxes[0].VarContent); i++ {
// 		wg.Add(1)
// 		go func(i int) {
// 			defer wg.Done()
// 			labeledImg := img
// 			var err error
// 			for _, tb := range textData.VarTextBoxes {
// 				labeledImg, err = addLabel(labeledImg, tb.MetaDetails, tb.Location, tb.VarContent[i])
// 				if err != nil {
// 					errors <- fmt.Errorf("error adding label: %w", err)
// 					return
// 				}
// 			}
// 			imgBytes, err := saveImage(labeledImg)
// 			if err != nil {
// 				errors <- fmt.Errorf("error saving image: %w", err)
// 				return
// 			}
// 			mu.Lock()
// 			imagesBytes = append(imagesBytes, imgBytes)
// 			mu.Unlock()
// 		}(i)
// 	}
// 	wg.Wait()
// 	close(errors)

// 	for err := range errors {
// 		if err != nil {
// 			return nil, err
// 		}
// 	}

//		log.Println("All variable images processed successfully.")
//		return imagesBytes, nil
//	}
func addImage(baseImg image.Image, img image.Image, location image.Point) image.Image {
	canvas := image.NewRGBA(baseImg.Bounds())
	draw.Draw(canvas, canvas.Bounds(), baseImg, image.Point{}, draw.Src)
	draw.Draw(canvas, baseImg.Bounds(), img, location, draw.Over)
	return canvas
}
func printVarContent(img image.Image, textData ishareRequest) ([][]byte, error) {
	var imagesBytes [][]byte // To store the bytes of all generated images
	if len(textData.VarTextBoxes) == 0 {
		imgBytes, err := saveImage(img)
		if err != nil {
			return nil, err
		}
		imagesBytes = append(imagesBytes, imgBytes)
		return imagesBytes, nil
	}

	var mu sync.Mutex                                                         // To safely append to imagesBytes
	errors := make(chan error, len(textData.VarTextBoxes[0].VarContent))      // Channel for errors
	imagesChan := make(chan []byte, len(textData.VarTextBoxes[0].VarContent)) // Channel for image bytes

	// Number of workers
	numWorkers := 10
	var wg sync.WaitGroup

	// Worker function
	worker := func(jobs <-chan int) {
		for i := range jobs {
			labeledImg := img
			var err error
			for _, tb := range textData.VarTextBoxes {
				labeledImg, err = addLabel(labeledImg, tb.MetaDetails, tb.Location, tb.VarContent[i])
				if err != nil {
					errors <- fmt.Errorf("error adding label: %w", err)
					return
				}
			}
			imgBytes, err := saveImage(labeledImg)
			if err != nil {
				errors <- fmt.Errorf("error saving image: %w", err)
				return
			}
			mu.Lock()
			imagesBytes = append(imagesBytes, imgBytes)
			mu.Unlock()
		}
		wg.Done()
	}

	// Create a channel to send jobs to workers
	jobs := make(chan int, len(textData.VarTextBoxes[0].VarContent))

	// Start workers
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go worker(jobs)
	}

	// Send jobs to workers
	for i := 0; i < len(textData.VarTextBoxes[0].VarContent); i++ {
		jobs <- i
	}
	close(jobs)

	// Wait for all workers to complete
	wg.Wait()
	close(imagesChan)
	close(errors)

	// Collect any errors
	for err := range errors {
		if err != nil {
			return nil, err
		}
	}

	log.Println("All variable images processed successfully.")
	return imagesBytes, nil
}
