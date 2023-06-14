package main

import (
	"fmt"

	"github.com/yomorun/yomo/serverless"
	"github.com/yomorun/yomo/serverless/guest"
)

func main() {
	guest.DataTags = DataTags
	guest.Handler = Handler
}

func Handler(ctx serverless.Context) {
	// load input data
	tag := ctx.Tag()
	input := ctx.Data()
	fmt.Printf("wasm go sfn received %d bytes with tag[%#x]\n", len(input), tag)

	// process app data
	// output := strings.ToUpper(string(input))
	var url string
	url = "https://example.org"
	retCode := ctx.HTTP().Get(url)
	fmt.Printf("wasm go sfn HTTPGet retCode=%d\n", retCode)

	// dump output data
	// ctx.Write(0x34, []byte(output))
}

func DataTags() []uint32 {
	return []uint32{0x33}
}
