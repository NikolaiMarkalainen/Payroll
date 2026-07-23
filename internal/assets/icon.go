package assets

import (
	_ "embed"

	"fyne.io/fyne/v2"
)

//go:embed icon.png
var iconPNG []byte

// AppIcon is the window / taskbar icon for Palkkatarkistus.
func AppIcon() fyne.Resource {
	return fyne.NewStaticResource("icon.png", iconPNG)
}
