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

func TestColor(t *testing.T) {

	function, _, _, _ := runtime.Caller(0)
	fn := runtime.FuncForPC(function).Name()
	var log = console.New()
	log.OK(fmt.Sprintf("%s\n", fn[strings.LastIndex(fn, ".Test")+5:]))
	a := assert.New(t)
	_ = a

	path := "temp/big.png"
	r, _ := NewFromFile(path)
	r.Resize(1200)
	file.Save("temp/big_resize1200.png", r.Export(90).Bytes())

	/* path := "temp/position.jpg"
	r, err := NewFromFile(path)
	a.Nil(err)

	r.SmartAvatar(200)
	file.Save("temp/position_avatar.jpg", r.JPG(80).Bytes()) */

	/* 	c, err := prominentcolor.Kmeans(r.Image)
	   	a.Nil(err)
	   	fmt.Println("#" + c[0].AsString())
	   	fmt.Println("#" + c[1].AsString())
	   	fmt.Println("#" + c[2].AsString())

	   	for _, count := range rankColorsCount(r.Colors())[:10] {
	   		fmt.Println(count.Value, "#"+Rgb2Hex(count.Key.RGBA()))
	   	} */

	//colors.RGBA()

	return

}

func color7(hex string) (colorname string) {

	//colors.RGBA()
	// HSL x, 100%, 50%
	return
}
