package images

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image/color"
	"sort"
)

func IsBase64(data []byte) bool {
	return bytes.Index(data, []byte(";base64,")) > -1
}

func Base64ToFile(data []byte) (res []byte, err error) {
	idx := bytes.Index(data, []byte(";base64,"))
	if idx < 0 {
		err = fmt.Errorf("wrong")
		return
	}
	reader := base64.NewDecoder(base64.StdEncoding, bytes.NewReader(data[idx+8:]))
	buff := bytes.Buffer{}
	_, err = buff.ReadFrom(reader)
	if err != nil {
		return
	}
	return buff.Bytes(), nil
}

/*
func FileToImage(file []byte) (img image.Image, width, height int, proportion float64, ext string, err error) {
	reader := bytes.NewReader(file)
	size, _, _ := image.DecodeConfig(reader)
	width = size.Width
	height = size.Height
	proportion = float64(width) / float64(height)
	img, ext, err = image.Decode(reader)
	return
} */

func rankColorsCount(colorsList map[color.Color]int) PairList {
	pl := make(PairList, len(colorsList))
	i := 0
	for k, v := range colorsList {
		pl[i] = Pair{k, v}
		i++
	}
	sort.Sort(sort.Reverse(pl))
	return pl
}

type Pair struct {
	Key   color.Color
	Value int
}

type PairList []Pair

func (p PairList) Len() int           { return len(p) }
func (p PairList) Less(i, j int) bool { return p[i].Value < p[j].Value }
func (p PairList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func Rgb2Hex(r, g, b uint32, a ...uint32) string {
	return fmt.Sprintf("%x%x%x", uint8(r/257), uint8(g/257), uint8(b/257))
}
