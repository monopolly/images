package images

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"

	"github.com/disintegration/imaging"
	"github.com/monopolly/errors"
	"github.com/monopolly/file"
	"github.com/muesli/smartcrop"
	"github.com/muesli/smartcrop/nfnt"
	"github.com/nxshock/colorcrop"
	"github.com/rwcarlsen/goexif/exif"
)

func NewFromFile(path string) (a *Image, err error) {
	b, err := file.Open(path)
	if err != nil {
		return
	}

	return New(b)
}

func NewFromFileE(path string) (a *Image, err error) {
	b, err := file.Open(path)
	if err != nil {
		return
	}
	return New(b)
}

/* даем обычный файл или base64 и получаем картинку с параметрами */
func New(img []byte) (a *Image, err errors.E) {

	if img == nil {
		return nil, errors.Empty()
	}

	image.RegisterFormat("jpeg", "jpeg", jpeg.Decode, jpeg.DecodeConfig)
	image.RegisterFormat("png", "png", png.Decode, png.DecodeConfig)
	image.RegisterFormat("gif", "gif", gif.Decode, gif.DecodeConfig)

	var er error
	if IsBase64(img) {
		img, er = Base64ToFile(img)
		if er != nil {
			err = errors.File()
			return
		}
	}

	reader := bytes.NewReader(img)

	a = new(Image)
	a.Source = img
	a.Size = len(img)

	a.Image, a.Ext, er = Decode(reader)
	if er != nil {
		err = errors.File()
		return
	}
	if a.Ext == "jpeg" {
		a.Ext = "jpg"
	}

	a.Width = a.Image.Bounds().Size().X
	a.Height = a.Image.Bounds().Size().Y

	a.Filter = imaging.CatmullRom
	a.Quality = 80

	/* iccreader := bytes.NewReader(img)
	icc, err := iccjpeg.GetICCBuf(iccreader)
	if err != nil {
		return
	}

	fmt.Println("has icc") */
	/* a.ICC = icc */

	return
}

/* даем обычный файл или base64 и получаем картинку с параметрами */
func NewFromImage(img image.Image) (a *Image) {
	a = new(Image)
	a.Width = img.Bounds().Size().X
	a.Height = img.Bounds().Size().Y
	a.Image = img
	a.Filter = imaging.CatmullRom
	return
}

/*
type Preset struct {
	MaxWidth   int
	MaxHeight  int
	Brightness float64
	Contrast   float64
	Sharpen    float64
	Quality    int
	Filter     imaging.ResampleFilter
}

var PresetInterior = Preset{
	MaxHeight: 1200,
	Contrast:  1.1,
	Sharpen:   1.1,
	Filter:    imaging.CatmullRom,
	Quality:   80,
} */

type Image struct {
	Source  []byte
	Image   image.Image
	Ext     string
	Width   int
	Height  int
	Quality int
	Size    int
	Filter  imaging.ResampleFilter
}

// проходит все пиксели
func (a *Image) Pixels(handler func(p image.Point, colors color.Color) (stop bool)) {
	for x := 0; x < a.Image.Bounds().Max.X; x++ {
		for y := 0; y < a.Image.Bounds().Max.Y; y++ {
			if handler(image.Pt(x, y), a.Image.At(x, y)) {
				return
			}
		}
	}
}

// убирает белые и прозрачные пиксели
func (a *Image) CropBackground() *Image {
	a.Image = colorcrop.Crop(
		a.Image,          // for source image
		a.Image.At(0, 0), // crop white border
		0.2)              // with 50% thresold
	a.Width = a.Image.Bounds().Size().X
	a.Height = a.Image.Bounds().Size().Y
	return a
}

func (a *Image) Compress() *Image {
	switch a.Ext {
	case "png":
		//a.Image = lossypng.Compress(a.Image, lossypng.NoConversion, 10)
	case "jpg":
		//mozjpegbin.Encode(w, m, o)
	}
	return a
}

func (a *Image) Colors() (list map[color.Color]int) {
	list = make(map[color.Color]int)
	a.Pixels(func(x image.Point, c color.Color) (stop bool) {
		list[c]++
		return
	})
	return
}

// убирает белые и прозрачные пиксели
func (a *Image) RemoveBackgroundDirty(presicion ...float64) *Image {
	value := 0.05
	if len(presicion) > 0 {
		value = presicion[0]
	}
	init := a.Image.At(0, 0)
	newimage := image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{a.Width, a.Height}})
	for x := 0; x < a.Width; x++ {
		for y := 0; y < a.Height; y++ {
			pixel := a.Image.At(x, y)
			diff := CompareColor(init, pixel)
			switch {
			case diff < value:
				newimage.Set(x, y, color.Transparent)
			default:
				newimage.Set(x, y, pixel)
			}
		}
	}
	a.Image = newimage
	return a
}

func (a *Image) Export(quality ...int) (file *bytes.Buffer) {
	//смотрим формат
	file = new(bytes.Buffer)
	switch a.Ext {
	case "jpg":
		q := a.Quality
		if len(quality) > 0 {
			q = quality[0]
		}
		jpeg.Encode(file, a.Image, &jpeg.Options{Quality: q})
	case "png":
		png.Encode(file, a.Image)
	case "gif":
		gif.Encode(file, a.Image, &gif.Options{})
	default:
		return
	}
	return
}

func (a *Image) GIF() (file *bytes.Buffer) {
	//смотрим формат
	file = new(bytes.Buffer)
	gif.Encode(file, a.Image, &gif.Options{})
	return
}

func (a *Image) PNG() (file *bytes.Buffer) {
	file = new(bytes.Buffer)
	png.Encode(file, a.Image)
	return
}

// Quality ranges from 1 to 100 inclusive
func (a *Image) JPG(quality ...int) (file *bytes.Buffer) {
	file = new(bytes.Buffer)
	q := a.Quality
	if len(quality) > 0 {
		q = quality[0]
	}
	jpeg.Encode(file, a.Image, &jpeg.Options{Quality: q})
	return
}

func (a *Image) Origin() (file []byte) {
	return a.Source
}

/* func (a *Image) Bytes(quality ...int) []byte {
	return a.Export(quality...).Bytes()
} */

func (a *Image) Base64(quality ...int) string {
	return base64.StdEncoding.EncodeToString(a.Export(quality...).Bytes())
}

func (a *Image) Base64HTML(quality ...int) string {
	return fmt.Sprintf("data:image/jpg;base64,%s", base64.StdEncoding.EncodeToString(a.JPG(quality...).Bytes()))
}

func (a *Image) Sharpen(v ...float64) *Image {
	sh := 0.5
	if len(v) > 0 {
		sh = v[0]
	}
	img := imaging.Sharpen(a.Image, sh)
	n := *a
	n.Image = img
	return &n
}

func (a *Image) Brightness(v ...float64) *Image {
	sh := 1.1
	if len(v) > 0 {
		sh = v[0]
	}
	img := imaging.AdjustBrightness(a.Image, sh)
	n := *a
	n.Image = img
	return &n
}

func (a *Image) Contrast(percents ...float64) *Image {
	sh := 1.1
	if len(percents) > 0 {
		sh = percents[0]
	}
	img := imaging.AdjustContrast(a.Image, sh)
	n := *a
	n.Image = img
	return &n
}

/* width, height */
func (a *Image) Resize(size ...int) *Image {
	var width, height int
	switch len(size) {
	case 0:
		return a
	case 1:
		return a.ResizeWidth(size[0])
	case 2:
		width = size[0]  //0
		height = size[1] //1200
		switch {
		case width == 0 && height > 0:
			return a.ResizeHeight(height)
		case width > 0 && height == 0:
			return a.ResizeWidth(width)
		case width > 0 && height > 0 && width == height:
			return a.Square(width)
		default:
		}
	default:
		return a
	}

	img := imaging.Thumbnail(a.Image, width, height, a.Filter)

	newimage := *a
	newimage.Image = img
	newimage.Width = width
	newimage.Height = height
	return &newimage
}

func (a *Image) Interior() *Image {
	a.Quality = 80
	a.Filter = imaging.CatmullRom
	return a.ResizeHeight(1600).Contrast(1.1).Sharpen(0.8)
}

func (a *Image) Instagram() *Image {
	img := imaging.Thumbnail(a.Image, 1000, 1000, a.Filter)
	n := *a
	n.Image = img
	n.Width = 1000
	n.Height = 1000
	return n.Contrast(1.1).Sharpen()
}

func (a *Image) ResizeWidth(width int) *Image {
	if width == a.Width {
		return a
	}
	p := float64(a.Width) / float64(a.Height) //800 / 600 = 1.3333
	height := int(float64(width) / p)

	img := imaging.Resize(a.Image, width, height, a.Filter)
	n := *a
	n.Image = img
	n.Width = width
	n.Height = height
	return &n
}

func (a *Image) ResizeHeight(height int) *Image {
	if height == a.Height {
		return a
	}
	p := float64(a.Height) / float64(a.Width) //800 / 600 = 1.3333
	width := int(float64(height) / p)
	img := imaging.Resize(a.Image, width, height, a.Filter)
	n := *a
	n.Image = img
	n.Width = width
	n.Height = height
	return &n
}

// Thumbnail
func (a *Image) Square(size int) *Image {
	img := imaging.Thumbnail(a.Image, size, size, a.Filter)
	n := *a
	n.Image = img
	n.Width = size
	n.Height = size
	return &n
}

// Decode is image.Decode handling orientation in EXIF tags if exists.
// Requires io.ReadSeeker instead of io.Reader.
func Decode(reader io.ReadSeeker) (image.Image, string, error) {
	img, fmt, err := image.Decode(reader)
	if err != nil {
		return img, fmt, err
	}
	reader.Seek(0, io.SeekStart)
	orientation := getOrientation(reader)
	switch orientation {
	case "1":
	case "2":
		img = imaging.FlipV(img)
	case "3":
		img = imaging.Rotate180(img)
	case "4":
		img = imaging.Rotate180(imaging.FlipV(img))
	case "5":
		img = imaging.Rotate270(imaging.FlipV(img))
	case "6":
		img = imaging.Rotate270(img)
	case "7":
		img = imaging.Rotate90(imaging.FlipV(img))
	case "8":
		img = imaging.Rotate90(img)
	}

	return img, fmt, err
}

func getOrientation(reader io.Reader) string {
	x, err := exif.Decode(reader)
	if err != nil {
		return "1"
	}
	if x != nil {
		orient, err := x.Get(exif.Orientation)
		if err != nil {
			return "1"
		}
		if orient != nil {
			return orient.String()
		}
	}

	return "1"
}

type subImager interface {
	SubImage(r image.Rectangle) image.Image
}

func (a *Image) SmartAvatar(size int) (i *Image) {
	analyzer := smartcrop.NewAnalyzer(nfnt.NewDefaultResizer())

	topCrop, err := analyzer.FindBestCrop(i.Image, size, size)
	if err != nil {
		fmt.Println(err)
		return
	}
	im, ok := i.Image.(subImager)
	if !ok {
		fmt.Println("smartcrop/fail", "im,ok := i.Image.(subImager)")
		return
	}
	i.Image = im.SubImage(topCrop)
	return a
}
