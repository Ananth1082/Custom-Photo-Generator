package models

import (
	"image"
	"image/color"
)

type IshareRequest struct {
	VarTextBoxes  []VarTextBox
	VarImageBoxes []VarImageBox
}

type TextDetails struct {
	Name     string
	IsItalic bool
	IsBold   bool
	Color    color.RGBA
	Size     int
}

type ImageDetails struct {
	Size struct {
		Width  int
		Height int
	}
}
type VarImageBox struct {
	MetaDetails ImageDetails
	ImageLink   []string
	Location    image.Point
}

type VarTextBox struct {
	MetaDetails TextDetails
	VarContent  []string
	Location    image.Point
}
