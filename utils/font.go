package utils

import (
	models "CustomPhotoGenerator/m-v0/Models"
	"fmt"
	"os"
	"sync"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
)

// Load font based on text details, with caching
var fontCache = sync.Map{}

func LoadFont(td models.TextDetails) (*truetype.Font, error) {
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
