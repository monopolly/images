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
	"strings"

	"github.com/disintegration/imaging"
	"github.com/gen2brain/avif"
	"github.com/monopolly/errors"
	"github.com/monopolly/file"
	"github.com/muesli/smartcrop"
	"github.com/muesli/smartcrop/nfnt"
	"github.com/nxshock/colorcrop"
	"golang.org/x/image/webp"
)

const (
	_ = int(iota)
	TypeJPG
	TypePNG
	TypeGIF
	TypeWebp
	TypeAvif
)

const (
	_ = int(iota)
	ShapeBox
	ShapePortrait
	ShaperLandscape
)

func init() {
	image.RegisterFormat("jpeg", "jpeg", jpeg.Decode, jpeg.DecodeConfig)
	image.RegisterFormat("png", "png", png.Decode, png.DecodeConfig)
	image.RegisterFormat("gif", "gif", gif.Decode, gif.DecodeConfig)
	image.RegisterFormat("webp", "webp", webp.Decode, webp.DecodeConfig)
	image.RegisterFormat("avif", "????ftypavif", avif.Decode, avif.DecodeConfig)
	image.RegisterFormat("avif", "????ftypavis", avif.Decode, avif.DecodeConfig)
}

func NewFromFile(path string) (a *Image, err error) {
	b, err := file.Open(path)
	if err != nil {
		return
	}

	return New(b)
}

func NewFromFileE(path string) (a *Image, err error) {
	return NewFromFile(path)
}

/* даем обычный файл или base64 и получаем картинку с параметрами */
func New(img []byte) (a *Image, err errors.E) {
	return newImage(img, Decode)
}

/* даем обычный файл или base64 и получаем картинку с параметрами (без учёта EXIF-ориентации) */
func NewWithoutExif(img []byte, _ ...bool) (a *Image, err errors.E) {
	return newImage(img, DecodeWithoutExif)
}

// newImage декодирует бинарную картинку или base64 и заполняет метаданные.
// decode задаёт способ декодирования (с учётом EXIF или без).
func newImage(img []byte, decode func(io.Reader) (image.Image, string, error)) (*Image, errors.E) {
	if len(img) == 0 {
		return nil, errors.Empty()
	}

	// Если это ещё не бинарная картинка — пробуем раскодировать base64,
	// но подставляем результат только когда после декода это действительно
	// картинка. Так валидный бинарник никогда не портится.
	if !IsImage(img) {
		if decoded, e := Base64ToFile(string(img)); e == nil && IsImage(decoded) {
			img = decoded
		}
	}

	a := new(Image)
	a.Source = img
	a.Size = len(img)

	var er error
	a.Image, a.Ext, er = decode(bytes.NewReader(img))
	if er != nil {
		return nil, errors.File("Cant decode image: " + er.Error())
	}

	a.applyMeta()
	return a, nil
}

// applyMeta заполняет формат, размеры, форму и параметры по умолчанию
// на основе уже декодированного a.Image и a.Ext.
func (a *Image) applyMeta() {
	switch a.Ext {
	case "jpg", "jpeg":
		a.Ext = "jpg"
		a.Type = TypeJPG
		a.Mime = "image/jpeg"
	case "png":
		a.Type = TypePNG
		a.Mime = "image/png"
	case "gif":
		a.Type = TypeGIF
		a.Mime = "image/gif"
	case "webp":
		a.Type = TypeWebp
		a.Mime = "image/webp"
	case "avif":
		a.Type = TypeAvif
		a.Mime = "image/avif"
	}

	a.Width = a.Image.Bounds().Dx()
	a.Height = a.Image.Bounds().Dy()

	switch {
	case a.Width == a.Height:
		a.Shape = ShapeBox
		a.Ratio = 1
	case a.Width > a.Height:
		a.Shape = ShaperLandscape
		a.Ratio = float64(a.Width) / float64(a.Height)
	default:
		a.Shape = ShapePortrait
		a.Ratio = float64(a.Height) / float64(a.Width)
	}

	a.Filter = imaging.CatmullRom
	a.Quality = 80
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

type Image struct {
	Source  []byte
	Image   image.Image
	Ext     string  //png,jpg
	Mime    string  //image/jpeg
	Type    int     //png,jpg
	Shape   int     //box,landscape
	Ratio   float64 //0.235,0.42114
	Width   int     //640
	Height  int     //480
	Quality int     //80
	Size    int     //245624 bytes
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

// brew install pngquant linux: https://pkgs.org/download/pngquant
func (a *Image) ExternalCompressPNG(level pngQuality) (res []byte) {
	res, _ = ExternalCompressPNG(a.PNG().Bytes(), level)
	return
}

// brew install jpegoptim quality 1-100
func (a *Image) ExternalCompressJPG(quality int) (res []byte) {
	res, _ = ExternalCompressJPG(a.JPG().Bytes(), quality)
	return
}

// brew install jpegoptim quality 1-100
func (a *Image) Compress() (res []byte) {
	switch a.Type {
	case TypeJPG:
		res = a.JPG(80).Bytes()
		b, _ := ExternalCompressJPG(res, 80)
		switch b {
		case nil:
			return
		default:
			return b
		}
	case TypePNG:
		res = a.PNG().Bytes()
		b, _ := ExternalCompressPNG(res, PNGQualityMediumCompression)
		switch b {
		case nil:
			return
		default:
			return b
		}
	default:
		return a.Export(80).Bytes()

	}

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
	case "webp":
		file, _ = Webp(a.Image)
	case "avif":
		q := 50
		if len(quality) > 0 {
			q = quality[0]
		}
		if b, err := Avifs(a.Image, q); err == nil {
			file = &b
		}
	default:
		jpeg.Encode(file, a.Image, &jpeg.Options{Quality: 80})
	}
	return
}

func (a *Image) GIF(options ...*gif.Options) (file *bytes.Buffer) {
	//смотрим формат
	file = new(bytes.Buffer)
	switch options != nil {
	case true:
		gif.Encode(file, a.Image, options[0])
	default:
		gif.Encode(file, a.Image, &gif.Options{})
	}

	return
}

func (a *Image) PNG() (file *bytes.Buffer) {
	file = new(bytes.Buffer)
	png.Encode(file, a.Image)
	return
}

// best for png only, something with transparency
func (a *Image) Webp() (res *bytes.Buffer) {
	res, _ = Webp(a.Image)
	return
}

// best for png only, something with transparency
func (a *Image) Avif(quality ...int) (res *bytes.Buffer) {
	q := 50
	if len(quality) > 0 {
		q = quality[0]
	}
	b, _ := Avifs(a.Image, q)
	return &b
}

func (a *Image) PNGBase64HTML() string {
	return fmt.Sprintf("data:image/png;base64,%s", base64.StdEncoding.EncodeToString(a.PNG().Bytes()))
}

func (a *Image) JPGBase64HTML(quality ...int) string {
	return fmt.Sprintf("data:image/jpeg;base64,%s", base64.StdEncoding.EncodeToString(a.JPG(quality...).Bytes()))
}

func (a *Image) GIFBase64HTML(options ...*gif.Options) string {
	return fmt.Sprintf("data:image/gif;base64,%s", base64.StdEncoding.EncodeToString(a.GIF(options...).Bytes()))
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

func (a *Image) Base64(quality ...int) string {
	return base64.StdEncoding.EncodeToString(a.Export(quality...).Bytes())
}

func (a *Image) Base64HTML(quality ...int) string {
	var res []byte
	var typ string
	switch a.Type {
	case TypeJPG:
		res = a.JPG(quality...).Bytes()
		typ = "jpeg"
	case TypePNG:
		res = a.PNG().Bytes()
		typ = "png"
	case TypeGIF:
		res = a.GIF().Bytes()
		typ = "gif"
	case TypeWebp:
		res = a.Webp().Bytes()
		typ = "webp"
	case TypeAvif:
		res = a.Avif().Bytes()
		typ = "avif"
	}
	return fmt.Sprintf("data:image/%s;base64,%s", typ, base64.StdEncoding.EncodeToString(res))
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

func (a *Image) Blur(v ...float64) *Image {
	sh := 0.5
	if len(v) > 0 {
		sh = v[0]
	}
	img := imaging.Blur(a.Image, sh)
	n := *a
	n.Image = img
	return &n
}

func (a *Image) Brightness(percents float64) *Image {
	img := imaging.AdjustBrightness(a.Image, percents)
	n := *a
	n.Image = img
	return &n
}

func (a *Image) Contrast(percents float64) *Image {
	img := imaging.AdjustContrast(a.Image, percents)
	n := *a
	n.Image = img
	return &n
}

/* width, height */
func (a *Image) ResizeOld(size ...int) *Image {
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

/* width, height */
func (a *Image) Resize(maxside int) *Image {

	if a.Width < maxside && a.Height < maxside {
		return a
	}

	filter := imaging.Cosine
	// fmt.Println(a.Width, a.Height, a.Size, a.Ratio)
	switch {
	case a.Width > a.Height:
		ratio := float64(a.Height) / float64(a.Width)
		a.Width = maxside
		a.Height = int(float64(maxside) * ratio)
		a.Image = imaging.Resize(a.Image, a.Width, a.Height, filter)
	case a.Width < a.Height:
		ratio := float64(a.Width) / float64(a.Height)
		a.Height = maxside
		a.Width = int(float64(maxside) * ratio)
		a.Image = imaging.Resize(a.Image, a.Width, a.Height, filter)
	default:
		a.Height = maxside
		a.Width = maxside
		a.Image = imaging.Resize(a.Image, a.Width, a.Height, filter)
	}

	return a
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
	if width > a.Width {
		return a
	}
	p := float64(a.Width) / float64(a.Height) //800 / 600 = 1.3333
	height := int(float64(width) / p)
	img := imaging.Resize(a.Image, width, height, a.Filter)
	a.Image = img
	a.Width = width
	a.Height = height
	return a
}

func (a *Image) ResizeHeight(height int) *Image {
	if height == a.Height {
		return a
	}
	p := float64(a.Height) / float64(a.Width) //800 / 600 = 1.3333
	width := int(float64(height) / p)
	img := imaging.Resize(a.Image, width, height, a.Filter)
	a.Image = img
	a.Width = width
	a.Height = height
	return a
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

type subImager interface {
	SubImage(r image.Rectangle) image.Image
}

func (a *Image) SmartAvatar(size int) *Image {
	analyzer := smartcrop.NewAnalyzer(nfnt.NewDefaultResizer())

	topCrop, err := analyzer.FindBestCrop(a.Image, size, size)
	if err != nil {
		fmt.Println(err)
		return a
	}
	im, ok := a.Image.(subImager)
	if !ok {
		fmt.Println("smartcrop/fail", "im,ok := a.Image.(subImager)")
		return a
	}
	a.Image = im.SubImage(topCrop)
	a.Width = a.Image.Bounds().Dx()
	a.Height = a.Image.Bounds().Dy()
	return a
}

func HTMLBase64(body []byte, imageType int) string {
	var typ string
	switch imageType {
	case TypePNG:
		typ = "png"
	case TypeGIF:
		typ = "gif"
	case TypeWebp:
		typ = "webp"
	case TypeAvif:
		typ = "avif"
	default:
		typ = "jpeg"
	}

	return fmt.Sprintf("data:image/%s;base64,%s", typ, base64.StdEncoding.EncodeToString(body))
}

func IsBase64(s string) bool {

	// Case 1: data URL
	if IsBase64ImageDataURL(s) {
		return true
	}

	// Case 2: raw base64
	return DecodeBase64(s) == nil
}

func IsBase64ImageDataURL(s string) bool {
	return strings.HasPrefix(s, "data:image/") &&
		strings.Contains(s, ";base64,")
}

func DecodeBase64(s string) (err error) {
	// remove whitespace/newlines
	s = strings.TrimSpace(s)

	// принимаем любой из поддерживаемых вариантов base64
	for _, enc := range base64Encodings {
		if _, err = enc.DecodeString(s); err == nil {
			return nil
		}
	}
	return
}

func IsImage(b []byte) bool {
	return DetectImageType(b) > 0
}

func DetectImageType(b []byte) (types int) {
	if len(b) < 12 {
		return
	}

	switch {
	case bytes.HasPrefix(b, []byte{0xFF, 0xD8, 0xFF}):
		return TypeJPG // true // JPEG
	case bytes.HasPrefix(b, []byte{0x89, 'P', 'N', 'G'}):
		return TypePNG // PNG
	case bytes.HasPrefix(b, []byte("GIF87a")),
		bytes.HasPrefix(b, []byte("GIF89a")):
		return TypeGIF // GIF
	case bytes.HasPrefix(b, []byte("RIFF")) && bytes.Contains(b[:12], []byte("WEBP")):
		return TypeWebp // WEBP
	case isAVIF(b):
		return TypeAvif
	default:
		return
	}
}

func isAVIF(b []byte) bool {
	if len(b) < 12 {
		return false
	}

	// ISO BMFF: size(4) + "ftyp"(4) + brand(4)
	return bytes.Equal(b[4:8], []byte("ftyp")) &&
		(bytes.Equal(b[8:12], []byte("avif")) ||
			bytes.Equal(b[8:12], []byte("avis")))
}

// remove data url data:image/
func CleanBase64(s string) string {
	s = strings.TrimSpace(s)

	const prefix = "data:image/"
	if !strings.HasPrefix(s, prefix) {
		return s
	}

	// ищем разделитель ";base64,"
	idx := strings.Index(s, ";base64,")
	if idx == -1 {
		return s
	}

	return s[idx+len(";base64,"):]
}
