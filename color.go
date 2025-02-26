package images

import (
	"image/color"
)

type Color struct {
	color.Color
	// Count int
}

func (a Color) Hex() string {
	return ColorToHex(a.Color)
}

func (a Color) HSL() (res HSL) {
	res.Hue, res.Saturation, res.Lightness = ColorToHSL(a.Color)
	return
}
