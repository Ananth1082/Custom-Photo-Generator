package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"

	"os"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	_ "golang.org/x/image/font"
	"sync"
)

type ishareRequest struct {
	MergedImage    string         
	ConstTextBoxes []ConstTextBox 
	VarTextBoxes   []VarTextBox   
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
func saveImage(img image.Image, filename string) error {
	outFile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer outFile.Close()

	return png.Encode(outFile, img)
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
func printConstContent(img image.Image, textData ishareRequest) (image.Image, error) {
	if len(textData.ConstTextBoxes) == 0 {
		return img, nil
	}

	for _, tb := range textData.ConstTextBoxes {
		var err error
		img, err = addLabel(img, tb.MetaDetails, tb.Location, tb.ConstContent)
		if err != nil {
			return nil, err
		}
	}
	err := saveImage(img, "./OUT/const_output.png")
	if err != nil {
		return nil, err
	}

	log.Println("Constant image saved successfully.")
	return img, nil
}

// Print variable content on the image
func printVarContent(img image.Image, textData ishareRequest) error {
	if len(textData.VarTextBoxes) == 0 {
		return saveImage(img, "./OUT/output.png")
	}

	var wg sync.WaitGroup
	for i := 0; i < len(textData.VarTextBoxes[0].VarContent); i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			labeledImg := img
			var err error
			for _, tb := range textData.VarTextBoxes {
				labeledImg, err = addLabel(labeledImg, tb.MetaDetails, tb.Location, tb.VarContent[i])
				if err != nil {
					log.Printf("An error occurred while adding label: %v", err)
					return
				}
			}
			err = saveImage(labeledImg, fmt.Sprintf("./OUT/output_%d.png", i))
			if err != nil {
				log.Printf("An error occurred while saving image: %v", err)
			}
		}(i)
	}
	wg.Wait()

	log.Println("All variable images saved successfully.")
	return nil
}
