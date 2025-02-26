package images

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"math"
	"strconv"
	"strings"
)

func (a *Image) Colors() (list map[color.Color]int) {
	list = make(map[color.Color]int)
	a.Pixels(func(x image.Point, c color.Color) (stop bool) {
		list[c]++
		return
	})
	return
}

// ColorToHSL convert color.Color into HSL triple, ignoring the alpha channel.
func ColorToHSL(c color.Color) (h, s, l float64) {
	r, g, b, _ := c.RGBA()
	return RGBToHSL(uint8(r>>8), uint8(g>>8), uint8(b>>8))
}

// ColorToHSV convert color.Color into HSV triple, ignoring the alpha channel.
func ColorToHSV(c color.Color) (h, s, v float64) {
	r, g, b, _ := c.RGBA()
	return RGBToHSV(uint8(r>>8), uint8(g>>8), uint8(b>>8))
}

// ColorToHex convert color.Color into Hex string, ignoring the alpha channel.
func ColorToHex(c color.Color) string {
	r, g, b, _ := c.RGBA()
	return RGBToHex(uint8(r>>8), uint8(g>>8), uint8(b>>8))
}

// HSLToColor convert HSL triple into color.Color.
func HSLToColor(h, s, l float64) (color.Color, error) {
	r, g, b, err := HSLToRGB(h, s, l)
	if err != nil {
		return nil, err
	}
	return color.RGBA{R: r, G: g, B: b, A: 0}, nil
}

// HSVToColor convert HSV triple into color.Color.
func HSVToColor(h, s, v float64) (color.Color, error) {
	r, g, b, err := HSVToRGB(h, s, v)
	if err != nil {
		return nil, err
	}
	return color.RGBA{R: r, G: g, B: b, A: 0}, nil
}

// HexToColor convert Hex string into color.Color.
func HexToColor(hex string) (color.Color, error) {
	r, g, b, err := HexToRGB(hex)
	if err != nil {
		return nil, err
	}
	return color.RGBA{R: r, G: g, B: b, A: 0}, nil
}

// RGBToHSL converts an RGB triple to an HSL triple.
func RGBToHSL(r, g, b uint8) (h, s, l float64) {
	// convert uint32 pre-multiplied value to uint8
	// The r,g,b values are divided by 255 to change the range from 0..255 to 0..1:
	Rnot := float64(r) / 255
	Gnot := float64(g) / 255
	Bnot := float64(b) / 255
	Cmax, Cmin := getMaxMin(Rnot, Gnot, Bnot)
	Δ := Cmax - Cmin
	// Lightness calculation:
	l = (Cmax + Cmin) / 2
	// Hue and Saturation Calculation:
	if Δ == 0 {
		h = 0
		s = 0
	} else {
		switch Cmax {
		case Rnot:
			h = 60 * (math.Mod((Gnot-Bnot)/Δ, 6))
		case Gnot:
			h = 60 * (((Bnot - Rnot) / Δ) + 2)
		case Bnot:
			h = 60 * (((Rnot - Gnot) / Δ) + 4)
		}
		if h < 0 {
			h += 360
		}

		s = Δ / (1 - math.Abs((2*l)-1))
	}

	return h, round(s), round(l)
}

// HSLToRGB converts an HSL triple to an RGB triple.
func HSLToRGB(h, s, l float64) (r, g, b uint8, err error) {
	if h < 0 || h >= 360 ||
		s < 0 || s > 1 ||
		l < 0 || l > 1 {
		return 0, 0, 0, fmt.Errorf("range")
	}
	// When 0 ≤ h < 360, 0 ≤ s ≤ 1 and 0 ≤ l ≤ 1:
	C := (1 - math.Abs((2*l)-1)) * s
	X := C * (1 - math.Abs(math.Mod(h/60, 2)-1))
	m := l - (C / 2)
	var Rnot, Gnot, Bnot float64

	switch {
	case 0 <= h && h < 60:
		Rnot, Gnot, Bnot = C, X, 0
	case 60 <= h && h < 120:
		Rnot, Gnot, Bnot = X, C, 0
	case 120 <= h && h < 180:
		Rnot, Gnot, Bnot = 0, C, X
	case 180 <= h && h < 240:
		Rnot, Gnot, Bnot = 0, X, C
	case 240 <= h && h < 300:
		Rnot, Gnot, Bnot = X, 0, C
	case 300 <= h && h < 360:
		Rnot, Gnot, Bnot = C, 0, X
	}
	r = uint8(math.Round((Rnot + m) * 255))
	g = uint8(math.Round((Gnot + m) * 255))
	b = uint8(math.Round((Bnot + m) * 255))
	return r, g, b, nil
}

// RGBToHSV converts an RGB triple to an HSV triple.
func RGBToHSV(r, g, b uint8) (h, s, v float64) {
	// convert uint32 pre-multiplied value to uint8
	// The r,g,b values are divided by 255 to change the range from 0..255 to 0..1:
	Rnot := float64(r) / 255
	Gnot := float64(g) / 255
	Bnot := float64(b) / 255
	Cmax, Cmin := getMaxMin(Rnot, Gnot, Bnot)
	Δ := Cmax - Cmin

	// Hue calculation:
	if Δ == 0 {
		h = 0
	} else {
		switch Cmax {
		case Rnot:
			h = 60 * (math.Mod((Gnot-Bnot)/Δ, 6))
		case Gnot:
			h = 60 * (((Bnot - Rnot) / Δ) + 2)
		case Bnot:
			h = 60 * (((Rnot - Gnot) / Δ) + 4)
		}
		if h < 0 {
			h += 360
		}

	}
	// Saturation calculation:
	if Cmax == 0 {
		s = 0
	} else {
		s = Δ / Cmax
	}
	// Value calculation:
	v = Cmax

	return h, round(s), round(v)
}

// HSVToRGB converts an HSV triple to an RGB triple.
func HSVToRGB(h, s, v float64) (r, g, b uint8, err error) {
	if h < 0 || h >= 360 ||
		s < 0 || s > 1 ||
		v < 0 || v > 1 {
		return 0, 0, 0, fmt.Errorf("range")
	}
	// When 0 ≤ h < 360, 0 ≤ s ≤ 1 and 0 ≤ v ≤ 1:
	C := v * s
	X := C * (1 - math.Abs(math.Mod(h/60, 2)-1))
	m := v - C
	var Rnot, Gnot, Bnot float64
	switch {
	case 0 <= h && h < 60:
		Rnot, Gnot, Bnot = C, X, 0
	case 60 <= h && h < 120:
		Rnot, Gnot, Bnot = X, C, 0
	case 120 <= h && h < 180:
		Rnot, Gnot, Bnot = 0, C, X
	case 180 <= h && h < 240:
		Rnot, Gnot, Bnot = 0, X, C
	case 240 <= h && h < 300:
		Rnot, Gnot, Bnot = X, 0, C
	case 300 <= h && h < 360:
		Rnot, Gnot, Bnot = C, 0, X
	}
	r = uint8(math.Round((Rnot + m) * 255))
	g = uint8(math.Round((Gnot + m) * 255))
	b = uint8(math.Round((Bnot + m) * 255))
	return r, g, b, nil
}

// RGBToHex converts an RGB triple to a Hex string in the format of 0xffff.
func RGBToHex(r, g, b uint8) string {
	// return fmt.Sprintf("0x%02x%02x%02x", r, g, b)
	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}

// HexToRGB converts a Hex string to an RGB triple.
func HexToRGB(hex string) (r, g, b uint8, err error) {
	// remove prefixes if found in the input string
	hex = strings.Replace(hex, "0x", "", -1)
	hex = strings.Replace(hex, "#", "", -1)
	if len(hex) != 6 {
		return 0, 0, 0, fmt.Errorf("hex")
	}

	r, err = hex2uint8(hex[0:2])
	if err != nil {
		return 0, 0, 0, err
	}
	g, err = hex2uint8(hex[2:4])
	if err != nil {
		return 0, 0, 0, err
	}
	b, err = hex2uint8(hex[4:6])
	if err != nil {
		return 0, 0, 0, err
	}
	return r, g, b, nil
}

// RGBToGrayAverage calculates the grayscale value of RGB with the average method, ignoring the alpha channel.
func RGBToGrayAverage(r, g, b uint8) color.Gray {
	return RGBToGrayWithWeight(r, g, b, 1, 1, 1)
}

// RGBToGrayWithWeight calculates the grayscale value of RGB wih provided weight, ignoring the alpha channel.
// In the standard library image/color, the conversion used the coefficient given by the JFIF specification. It is
// equivalent to using the weight 299, 587, 114 for rgb.
func RGBToGrayWithWeight(r, g, b uint8, rWeight, gWeight, bWeight uint) color.Gray {
	rw := uint(r) * rWeight
	gw := uint(g) * gWeight
	bw := uint(b) * bWeight

	return color.Gray{Y: uint8(math.Round(float64(rw+gw+bw) / float64(rWeight+gWeight+bWeight)))}
}

func hex2uint8(hexStr string) (uint8, error) {
	// base 16 for hexadecimal
	result, err := strconv.ParseUint(hexStr, 16, 8)
	if err != nil {
		return 0, err
	}
	return uint8(result), nil
}

func getMaxMin(a, b, c float64) (max, min float64) {
	if a > b {
		max = a
		min = b
	} else {
		max = b
		min = a
	}
	if c > max {
		max = c
	} else if c < min {
		min = c
	}
	return max, min
}

func round(x float64) float64 {
	return math.Round(x*1000) / 1000
}

// how many colors in dominant
// func (a *Image) DominantColors(count int) (list []*Color) {

// 	for k, v := range a.Colors() {
// 		list = append(list, &Color{Color: k, Count: v})
// 	}

// 	sort.Slice(list, func(i, j int) bool { return list[i].Count > list[j].Count })

// 	if len(list) > count {
// 		list = list[:count]
// 	}

// 	return
// }

type RGB struct {
	R, G, B float64
}

// Takes a string like '#123456' or 'ABCDEF' and returns an RGB
func HTMLToRGB(in string) (RGB, error) {
	if in[0] == '#' {
		in = in[1:]
	}

	if len(in) != 6 {
		return RGB{}, errors.New("Invalid string length")
	}

	var r, g, b byte
	if n, err := fmt.Sscanf(in, "%2x%2x%2x", &r, &g, &b); err != nil || n != 3 {
		return RGB{}, err
	}

	return RGB{float64(r) / 255, float64(g) / 255, float64(b) / 255}, nil
}

func (c RGB) ToHSL() HSL {
	var h, s, l float64

	r := c.R
	g := c.G
	b := c.B

	max := math.Max(math.Max(r, g), b)
	min := math.Min(math.Min(r, g), b)

	// Luminosity is the average of the max and min rgb color intensities.
	l = (max + min) / 2

	// saturation
	delta := max - min
	if delta == 0 {
		// it's gray
		return HSL{0, 0, l}
	}

	// it's not gray
	if l < 0.5 {
		s = delta / (max + min)
	} else {
		s = delta / (2 - max - min)
	}

	// hue
	r2 := (((max - r) / 6) + (delta / 2)) / delta
	g2 := (((max - g) / 6) + (delta / 2)) / delta
	b2 := (((max - b) / 6) + (delta / 2)) / delta
	switch {
	case r == max:
		h = b2 - g2
	case g == max:
		h = (1.0 / 3.0) + r2 - b2
	case b == max:
		h = (2.0 / 3.0) + g2 - r2
	}

	// fix wraparounds
	switch {
	case h < 0:
		h += 1
	case h > 1:
		h -= 1
	}

	return HSL{h, s, l}
}

// A nudge to make truncation round to nearest number instead of flooring
const delta = 1 / 512.0

func (c RGB) Hex() string {
	return fmt.Sprintf("#%02x%02x%02x", byte((c.R+delta)*255), byte((c.G+delta)*255), byte((c.B+delta)*255))
}

func hueToRGB(v1, v2, h float64) float64 {
	if h < 0 {
		h += 1
	}
	if h > 1 {
		h -= 1
	}
	switch {
	case 6*h < 1:
		return (v1 + (v2-v1)*6*h)
	case 2*h < 1:
		return v2
	case 3*h < 2:
		return v1 + (v2-v1)*((2.0/3.0)-h)*6
	}
	return v1
}
