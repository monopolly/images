package images

import (
	"fmt"

	"github.com/EdlinOrg/prominentcolor"
)

// dominant colors in hex
func (a *Image) DominantColorsHex(count int) (colors []string) {
	list, err := prominentcolor.Kmeans(a.Image)
	if err != nil {
		return
	}

	for _, x := range list {
		colors = append(colors, fmt.Sprintf("#%s", x.AsString()))
	}

	if len(colors) < count {
		return
	}

	return colors[:count]

}

// dominant colors in hex
func (a *Image) DominantColors(count int) (colors []*Color) {
	list, err := prominentcolor.Kmeans(a.Image)
	if err != nil {
		return
	}

	for _, x := range list {
		c, _ := HexToColor(x.AsString())
		colors = append(colors, &Color{Color: c})
	}

	if len(colors) < count {
		return
	}

	return colors[:count]

}

// most light from dominant colors in hex
func (a *Image) DominantColorLight() (color *Color) {
	var best float64
	for _, x := range a.DominantColors(2) {
		if l := x.HSL().Lightness; color == nil || l > best {
			color = x
			best = l
		}
	}
	return
}
