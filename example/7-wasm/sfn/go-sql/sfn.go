package main

import (
	"fmt"
	"log"

	"github.com/yomorun/yomo/serverless"
	"github.com/yomorun/yomo/serverless/guest"
)

func main() {
	guest.DataTags = DataTags
	guest.Handler = Handler
	// db open
	driverName := "postgres"
	dataSourceName := "postgresql://postgres:123456@localhost/test?sslmode=disable"
	err := guest.SQL.Open(driverName, dataSourceName)
	if err != nil {
		fmt.Printf("wasm go sfn SQL Open err=%v\n", err)
		return
	}
	log.Println("wasm go sfn SQL Open success")
}

func Handler(ctx serverless.Context) {
	// load input data
	tag := ctx.Tag()
	input := ctx.Data()
	fmt.Printf("wasm go sfn received %d bytes with tag[%#x]\n", len(input), tag)

	// process app data
	// output := strings.ToUpper(string(input))
	// db query from zipper table
	// log.Printf("wasm go sfn db[%T]: %v\n", db, db)
	result, err := ctx.SQL().Query("SELECT * FROM message")
	if err != nil {
		fmt.Printf("wasm go sfn SQL Query err=%v\n", err)
		return
	}
	log.Printf("wasm go sfn result: %v\n", result)

	// dump output data
	// ctx.Write(0x34, []byte(output))
}

func DataTags() []uint32 {
	return []uint32{0x33}
}

type Zipper struct {
	Name string `json:"name"`
	Host string `json:"host"`
}
