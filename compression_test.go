package images

//testing

import (
	"fmt"
	"runtime"
	"strings"
	"testing"

	"github.com/monopolly/console"
	"github.com/monopolly/file"
	"github.com/stretchr/testify/assert"
)

func TestCompressionPNG(t *testing.T) {

	function, _, _, _ := runtime.Caller(0)
	fn := runtime.FuncForPC(function).Name()
	var log = console.New()
	log.OK(fmt.Sprintf("%s\n", fn[strings.LastIndex(fn, ".Test")+5:]))
	a := assert.New(t)
	_ = a

	path := "temp/big.png"
	path2 := "temp/big_compressed.png"
	v := file.OpenE(path)
	c, err := CompressPNG(v, BestCompression)
	if err != nil {
		panic(err)
	}
	file.Save(path2, c)

}
func TestCompressionJPG(t *testing.T) {

	function, _, _, _ := runtime.Caller(0)
	fn := runtime.FuncForPC(function).Name()
	var log = console.New()
	log.OK(fmt.Sprintf("%s\n", fn[strings.LastIndex(fn, ".Test")+5:]))
	a := assert.New(t)
	_ = a

	path := "temp/big.jpg"
	path2 := "temp/big_compressed2.jpg"
	v := file.OpenE(path)
	c, err := CompressJPG(v, 70)
	if err != nil {
		panic(err)
	}
	file.Save(path2, c)

}
