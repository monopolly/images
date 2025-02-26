package images

import (
	"bytes"
	"image"

	"github.com/HugoSmits86/nativewebp"
)

func Webp(img image.Image) (res *bytes.Buffer, err error) {
	res = new(bytes.Buffer)
	err = nativewebp.Encode(res, img, nil)
	if err != nil {
		return
	}
	return
}
