package widgets

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

func NewTodoListItem() fyne.CanvasObject {
	checkbox := widget.NewCheck("", func(value bool) {
		fmt.Println("Checked", value)
	})
	description := canvas.NewText("", color.White)
	content := container.New(layout.NewHBoxLayout(), checkbox, layout.NewSpacer(), description)
	return content
}
