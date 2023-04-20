package images

import (
	"encoding/base64"
	"fmt"
)

type Format struct {
	image *Image
	ext   int
}

// base64 no header
func (a *Format) Raw64(quality ...int) string {
	return base64.StdEncoding.EncodeToString(a.image.Export(quality...).Bytes())
}

// for html images
func (a *Format) Base64(quality ...int) string {
	var res []byte
	var ext string
	switch a.ext {
	case TypePNG:
		ext = "png"
		res = a.image.PNG().Bytes()
	case TypeGIF:
		ext = "gif"
		res = a.image.GIF().Bytes()
	default:
		ext = "jpg"
		res = a.image.JPG(quality...).Bytes()
	}
	return fmt.Sprintf("data:image/%s;base64,%s", ext, base64.StdEncoding.EncodeToString(res))
}

func (a *Format) PNG() *Format {
	a.ext = TypePNG
	return a
}
