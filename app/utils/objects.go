package utils

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
)

func NewSectionLabel(text string) fyne.CanvasObject {
	obj := canvas.NewText("Add a new todo:", color.White)
	obj.TextSize = 15
	obj.TextStyle = fyne.TextStyle{Bold: true}
	return obj
}
