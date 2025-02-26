package images

type HSL struct {
	Hue        float64
	Saturation float64
	Lightness  float64
}

func (c HSL) ToRGB() RGB {
	h := c.Hue
	s := c.Saturation
	l := c.Lightness

	if s == 0 {
		// it's gray
		return RGB{l, l, l}
	}

	var v1, v2 float64
	if l < 0.5 {
		v2 = l * (1 + s)
	} else {
		v2 = (l + s) - (s * l)
	}

	v1 = 2*l - v2

	r := hueToRGB(v1, v2, h+(1.0/3.0))
	g := hueToRGB(v1, v2, h)
	b := hueToRGB(v1, v2, h-(1.0/3.0))

	return RGB{r, g, b}
}

func (c HSL) Hex() string {
	return c.ToRGB().Hex()
}
