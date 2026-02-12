package images

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"image/color"
	"io"
	"net/http"
	"sort"
	"strings"
)

// от base64 bombs
const maxImageBytes = 20 << 20 // 20 MB

func DetectImageMIME(b []byte) (string, bool) {
	if len(b) == 0 {
		return "", false
	}

	head := b
	if len(head) > 512 {
		head = head[:512]
	}
	m := http.DetectContentType(head)

	if strings.HasPrefix(m, "image/") {
		return m, true
	}

	// AVIF / HEIF
	if len(b) >= 12 && bytes.Equal(b[4:8], []byte("ftyp")) {
		switch string(b[8:12]) {
		case "avif", "avis":
			return "image/avif", true
		case "heic", "heix", "hevc", "hevx", "mif1", "msf1":
			return "image/heif", true
		}
	}

	return "", false
}

func NormalizeImageBytes(input []byte) ([]byte, error) {
	if len(input) == 0 {
		return nil, errors.New("empty input")
	}

	// 1️⃣ Если это уже бинарная картинка — возвращаем как есть
	if _, ok := DetectImageMIME(input); ok {
		return input, nil
	}

	// 2️⃣ Иначе считаем, что это base64-текст
	s := strings.TrimSpace(string(input))

	// срезаем data:image/...;base64,
	s = StripDataImageBase64Prefix(s)

	// убираем пробелы / переводы строк
	s = strings.Map(func(r rune) rune {
		switch r {
		case ' ', '\n', '\r', '\t':
			return -1
		default:
			return r
		}
	}, s)

	// быстрый size check
	if len(s)*3/4 > maxImageBytes {
		return nil, errors.New("decoded image exceeds size limit")
	}

	// пробуем разные base64 encoding
	encs := []*base64.Encoding{
		base64.StdEncoding,
		base64.RawStdEncoding,
		base64.URLEncoding,
		base64.RawURLEncoding,
	}

	var decoded []byte
	var err error

	for _, enc := range encs {
		decoded, err = enc.DecodeString(s)
		if err == nil {
			break
		}
	}
	if err != nil {
		return nil, errors.New("invalid base64 image")
	}

	// 3️⃣ Проверяем, что после декода это картинка
	// if _, ok := DetectImageMIME(decoded); !ok {
	// 	return nil, errors.New("decoded data is not an image")
	// }

	return decoded, nil
}

func StripDataImageBase64Prefix(s string) string {
	s = strings.TrimSpace(s)

	const p = "data:image/"
	if !strings.HasPrefix(s, p) {
		return s
	}
	i := strings.Index(s, ";base64,")
	if i == -1 {
		return s
	}
	return s[i+len(";base64,"):]
}

func DecodeBase64Image(input string, maxDecodedBytes int64) ([]byte, error) {
	s := StripDataImageBase64Prefix(input)

	// убираем пробелы/переносы (часто base64 форматируют)
	s = strings.Map(func(r rune) rune {
		switch r {
		case ' ', '\n', '\r', '\t':
			return -1
		default:
			return r
		}
	}, strings.TrimSpace(s))

	// быстрый пред-чек, чтобы не пытаться декодировать гигабайты
	// оценка: decoded ~= len(s)*3/4
	if int64(len(s))*3/4 > maxDecodedBytes {
		return nil, fmt.Errorf("image too large: estimated decoded size exceeds limit")
	}

	// пробуем разные варианты base64
	encodings := []*base64.Encoding{
		base64.StdEncoding,
		base64.RawStdEncoding,
		base64.URLEncoding,
		base64.RawURLEncoding,
	}

	var lastErr error
	for _, enc := range encodings {
		r := base64.NewDecoder(enc, strings.NewReader(s))

		// жестко ограничиваем реальный вывод
		lr := &io.LimitedReader{R: r, N: maxDecodedBytes + 1}

		var buf bytes.Buffer
		if _, err := buf.ReadFrom(lr); err != nil {
			lastErr = err
			continue
		}
		if lr.N <= 0 {
			return nil, fmt.Errorf("image too large: decoded size exceeds limit")
		}

		out := buf.Bytes()
		if len(out) == 0 {
			lastErr = fmt.Errorf("empty decoded data")
			continue
		}
		return out, nil
	}

	if lastErr == nil {
		lastErr = fmt.Errorf("invalid base64")
	}
	return nil, lastErr
}

func Base64ToFile(data string) (res []byte, err error) {

	return NormalizeImageBytes([]byte(data))

	// data = StripDataImageBase64Prefix(data)

	// reader := base64.NewDecoder(base64.StdEncoding, strings.NewReader(data))
	// buff := bytes.Buffer{}
	// _, err = buff.ReadFrom(reader)
	// if err != nil {
	// 	return Base64ToFileRaw(data)
	// }
	// return buff.Bytes(), nil
}

// func Base64ToFileRaw(data string) (res []byte, err error) {
// 	data = CleanBase64(data)

// 	reader := base64.NewDecoder(base64.RawStdEncoding, strings.NewReader(data))
// 	buff := bytes.Buffer{}
// 	_, err = buff.ReadFrom(reader)
// 	if err != nil {
// 		return
// 	}
// 	return buff.Bytes(), nil
// }

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
