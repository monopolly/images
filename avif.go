package images

import (
	"bytes"
	"image"
	"log"

	"github.com/gen2brain/avif"
)

func Avifs(img image.Image, quality int) (buf bytes.Buffer, err error) {
	err = avif.Encode(&buf, img, avif.Options{
		Quality: quality, // 0-100
		Speed:   8,       // 0-10 (0=медленно/лучше)
	})
	return
}

// normilize, quality 50, speed 8
func Avif(img image.Image) []byte {
	b, err := Avifs(img, 50)
	if err != nil {
		log.Println("avif converter", err)
		return nil
	}
	return b.Bytes()
}
