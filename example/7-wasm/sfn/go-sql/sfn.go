package main

import (
	"database/sql"
	"errors"
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
		log.Printf("[SFN] database open error: %v\n", err)
		return
	}
	log.Println("[SFN] database open success")
}

func Handler(ctx serverless.Context) {
	log.Printf("[SFN] ----------------------------------------------------------\n")
	// load input data
	tag := ctx.Tag()
	input := ctx.Data()
	log.Printf("[SFN] received %d bytes with tag[%#x]\n", len(input), tag)

	// process app data
	// output := strings.ToUpper(string(input))

	// query
	result, err := ctx.SQL().
		Query("SELECT * FROM message Where id=$1 or id=$2 or msg=$3", 2, 4, "eee")
	if err != nil {
		log.Printf("[SFN] execute query error: %v\n", err)
		return
	}
	log.Printf("[SFN] execute query result: %v\n", result)

	// query row
	item, err := ctx.SQL().QueryRow("select * from message where id=$1", 6)
	if err != nil {
		log.Printf("[SFN] execute query row error: %v\n", err)
		return
	}
	log.Printf("[SFN] execute query row result: %v\n", item)
	// query row none
	item, err = ctx.SQL().QueryRow("select * from message where id=2")
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Println("[SFN] execute query row result: no rows")
		} else {
			log.Printf("[SFN] execute query row error: %v\n", err)
			return
		}
	} else {
		log.Printf("[SFN] execute query row result: %v\n", item)
	}

	// exec
	// insert
	sqlResult, err := ctx.SQL().Exec("insert into message(msg) values($1)", "insert row")
	if err != nil {
		log.Printf("[SFN] execute insert error: %v\n", err)
		return
	}
	log.Printf("[SFN] execute insert result: %+v\n", sqlResult)
	// update
	sqlResult, err = ctx.SQL().Exec("update message set msg=$1 where id=$2", "update row", 1)
	if err != nil {
		log.Printf("[SFN] execute update error: %v\n", err)
		return
	}
	log.Printf("[SFN] execute update result: %+v\n", sqlResult)
	// delete
	sqlResult, err = ctx.SQL().Exec("delete from message where id=2")
	if err != nil {
		log.Printf("[SFN] execute delete error: %v\n", err)
		return
	}
	log.Printf("[SFN] execute delete result: %+v\n", sqlResult)

	// dump output data
	// ctx.Write(0x34, []byte(output))
}

func DataTags() []uint32 {
	return []uint32{0x33}
}
