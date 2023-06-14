package wazero

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"

	_ "github.com/lib/pq"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	wasmsql "github.com/yomorun/yomo/cli/serverless/wasm/sql"
)

var DB *sql.DB

func ExportSQLHostFuncs(builder wazero.HostModuleBuilder) {
	builder.
		// open
		NewFunctionBuilder().
		WithGoModuleFunction(
			api.GoModuleFunc(Open),
			[]api.ValueType{api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32},
			[]api.ValueType{api.ValueTypeI32},
		).
		Export(wasmsql.WasmFuncSQLOpen).
		// query
		NewFunctionBuilder().
		WithGoModuleFunction(
			api.GoModuleFunc(Query),
			[]api.ValueType{api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32},
			[]api.ValueType{api.ValueTypeI32},
		).
		Export(wasmsql.WasmFuncSQLQuery)
}

// func Open(driverName, dataSourceName string) (*DB, error) {
func Open(ctx context.Context, m api.Module, stack []uint64) {
	// driver name
	driverNamePtr := uint32(stack[0])
	driverNameSize := uint32(stack[1])
	buf, ok := m.Memory().Read(driverNamePtr, driverNameSize)
	if !ok {
		log.Printf(
			"[SQLOpen] driver name: Memory.Read(%d, %d) out of range\n",
			driverNamePtr,
			driverNameSize,
		)
		stack[0] = 1
		return
	}
	driverName := make([]byte, driverNameSize)
	copy(driverName, buf)
	// log.Printf("[SQLOpen] driver name: %s\n", driverName)
	// data source name
	dataSourceNamePtr := uint32(stack[2])
	dataSourceNameSize := uint32(stack[3])
	buf, ok = m.Memory().Read(dataSourceNamePtr, dataSourceNameSize)
	if !ok {
		log.Printf(
			"[SQLOpen] data soruce name: Memory.Read(%d, %d) out of range\n",
			dataSourceNamePtr,
			dataSourceNameSize,
		)
		stack[0] = 2
		return
	}
	dataSourceName := make([]byte, dataSourceNameSize)
	copy(dataSourceName, buf)
	// log.Printf("[SQLOpen] data source name: %s\n", dataSourceName)
	// open
	db, err := sql.Open(string(driverName), string(dataSourceName))
	if err != nil {
		log.Printf("[SQLOpen] Open(%s, %s) error: %v\n", driverName, dataSourceName, err)
		stack[0] = 3
		return
	}
	// ping
	if err := db.Ping(); err != nil {
		log.Printf("[SQLOpen] Ping() error: %v\n", err)
		stack[0] = 4
		return
	}
	log.Println("[SQLOpen] Ping() success")
	// return db
	DB = db
	stack[0] = 0
	// dbPtr := uintptr(unsafe.Pointer(db))
	// stack[0] = uint64(dbPtr)
	// // m.Memory().WriteUint32Le(uint32(stack[0]), uint32(dbPtr))
	// log.Printf("[SQLOpen] DB[%T]: %v, stack[0]=%d, dbPtr=%d\n", db, db, stack[0], dbPtr)
	return
}

func Query(ctx context.Context, m api.Module, stack []uint64) {
	queryPtr := uint32(stack[0])
	querySize := uint32(stack[1])
	buf, ok := m.Memory().Read(queryPtr, querySize)
	if !ok {
		log.Printf(
			"[SQLQuery] query: Memory.Read(%d, %d) out of range\n",
			queryPtr,
			querySize,
		)
		stack[0] = 1
		return
	}
	query := make([]byte, querySize)
	copy(query, buf)
	log.Printf("[SQLQuery] query: %s\n", query)
	// query
	rows, err := DB.QueryContext(ctx, string(query))
	if err != nil {
		log.Printf("[SQLQuery] Query(%s) error: %v\n", query, err)
		stack[0] = 2
		return
	}
	defer rows.Close()
	result, err := rows2maps(rows)
	if err != nil {
		log.Printf("[SQLQuery] rows2maps() error: %v\n", err)
		stack[0] = 3
		return
	}
	log.Printf("[SQLQuery] result: %v\n", result)
	resultBuf, err := json.Marshal(result)
	if err != nil {
		log.Printf("[SQLQuery] Marshal() error: %v\n", err)
		stack[0] = 4
		return
	}

	if len(resultBuf) == 0 {
		log.Println("[SQLQuery] resultBuf is empty")
		stack[0] = 5
		return
	}
	log.Printf("[SQLQuery] resultBuf: %s\n", resultBuf)
	// allocate buffer and write to memory
	resultPtr := uint32(stack[2])
	resultSize := uint32(stack[3])
	log.Printf(
		"[SQLQuery] resultPtr=%d, resultSize=%d, bufLen=%d\n",
		resultPtr,
		resultSize,
		len(resultBuf),
	)
	if err := allocateBuffer(ctx, m, resultPtr, resultSize, resultBuf); err != nil {
		log.Printf("[SQLQuery] allocateBuffer() error: %v\n", err)
		stack[0] = 9
		return
	}
	stack[0] = 0
}

// rows2maps convert sql.Rows to []map[string]any
func rows2maps(rows *sql.Rows) (result []map[string]any, err error) {
	if rows == nil {
		return
	}
	columns, err := rows.Columns()
	if err != nil {
		return result, err
	}
	length := len(columns)
	values := make([]any, length)
	for i := 0; i < length; i++ {
		values[i] = new(any)
	}
	for rows.Next() {
		err = rows.Scan(values...)
		if nil != err {
			log.Printf("[rows2map] Scan() error: %v\n", err)
			return result, err
		}
		row := make(map[string]any)
		for i, name := range columns {
			row[name] = *(values[i].(*any))
		}
		result = append(result, row)
	}
	return
}
