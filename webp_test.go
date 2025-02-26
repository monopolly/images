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

func TestWebp(t *testing.T) {

	function, _, _, _ := runtime.Caller(0)
	fn := runtime.FuncForPC(function).Name()
	var log = console.New()
	log.OK(fmt.Sprintf("%s\n", fn[strings.LastIndex(fn, ".Test")+5:]))
	a := assert.New(t)
	_ = a

	path := "temp/big.png"
	r, _ := NewFromFile(path)

	res := r.Webp()
	file.Save("temp/big.webp", res.Bytes())

}
