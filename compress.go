package images

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

type pngQuality int

const (
	PNGQualityBestCompression   = pngQuality(1)
	PNGQualityMediumCompression = pngQuality(10)
	PNGQualityFastCompression   = pngQuality(20)
)

// brew install pngquant linux: https://pkgs.org/download/pngquant
func ExternalCompressPNG(v []byte, level pngQuality) (res []byte, err error) {
	cmd := exec.Command("pngquant", "-", "--strip", "--speed", fmt.Sprint(level))
	cmd.Stdin = strings.NewReader(string(v))
	var o bytes.Buffer
	cmd.Stdout = &o
	err = cmd.Run()
	if err != nil {
		return
	}

	return o.Bytes(), nil
}

// brew install jpegoptim
// func CompressJPG(v []byte, level pngQuality) (res []byte, err error) {
// 	cmd := exec.Command("jpegoptim", "-f", "-s", "--stdin", "--stdout", "--all-progressive", "--max="+fmt.Sprint(level))
// 	cmd.Stdin = strings.NewReader(string(v))
// 	var o bytes.Buffer
// 	cmd.Stdout = &o
// 	err = cmd.Run()
// 	if err != nil {
// 		return
// 	}

// 	return o.Bytes(), nil
// }

// brew install jpegoptim quality 1-100
func ExternalCompressJPG(v []byte, quality int) (res []byte, err error) {
	cmd := exec.Command("jpegoptim", "-f", "-s", "--stdin", "--stdout", "--all-progressive", "--max="+fmt.Sprint(quality))
	cmd.Stdin = strings.NewReader(string(v))
	var o bytes.Buffer
	cmd.Stdout = &o
	err = cmd.Run()
	if err != nil {
		return
	}

	return o.Bytes(), nil
}
