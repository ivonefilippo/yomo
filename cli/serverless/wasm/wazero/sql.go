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
			[]api.ValueType{
				api.ValueTypeI32, // queryPtr
				api.ValueTypeI32, // querySize
				api.ValueTypeI32, // argsPtr
				api.ValueTypeI32, // argsSize
				api.ValueTypeI32, // resultPtr
				api.ValueTypeI32, // resultSize
			},
			[]api.ValueType{api.ValueTypeI32},
		).
		Export(wasmsql.WasmFuncSQLQuery).
		// query row
		NewFunctionBuilder().
		WithGoModuleFunction(
			api.GoModuleFunc(QueryRow),
			[]api.ValueType{
				api.ValueTypeI32, // queryPtr
				api.ValueTypeI32, // querySize
				api.ValueTypeI32, // argsPtr
				api.ValueTypeI32, // argsSize
				api.ValueTypeI32, // resultPtr
				api.ValueTypeI32, // resultSize
			},
			[]api.ValueType{api.ValueTypeI32},
		).
		Export(wasmsql.WasmFuncSQLQueryRow).
		// exec
		NewFunctionBuilder().
		WithGoModuleFunction(
			api.GoModuleFunc(Exec),
			[]api.ValueType{
				api.ValueTypeI32, // queryPtr
				api.ValueTypeI32, // querySize
				api.ValueTypeI32, // argsPtr
				api.ValueTypeI32, // argsSize
				api.ValueTypeI32, // resultPtr
				api.ValueTypeI32, // resultSize
			},
			[]api.ValueType{api.ValueTypeI32},
		).
		Export(wasmsql.WasmFuncSQLExec).
		// close
		NewFunctionBuilder().
		WithGoModuleFunction(
			api.GoModuleFunc(Close),
			[]api.ValueType{},
			[]api.ValueType{api.ValueTypeI32},
		).
		Export(wasmsql.WasmFuncSQLClose)
}

// Open opens a database specified by its database driver name and a driver-specific data source name
func Open(ctx context.Context, m api.Module, stack []uint64) {
	// driver name
	driverNamePtr := uint32(stack[0])
	driverNameSize := uint32(stack[1])
	driverName, err := readBuffer(ctx, m, driverNamePtr, driverNameSize)
	if err != nil {
		log.Printf("[SQL] Open: get driver name error: %s\n", err)
		stack[0] = 1
		return
	}
	// data source name
	dataSourceNamePtr := uint32(stack[2])
	dataSourceNameSize := uint32(stack[3])
	dataSourceName, err := readBuffer(ctx, m, dataSourceNamePtr, dataSourceNameSize)
	if err != nil {
		log.Printf("[SQL] Open: get data soruce name error: %s\n", err)
		stack[0] = 2
		return
	}
	// open
	db, err := sql.Open(string(driverName), string(dataSourceName))
	if err != nil {
		log.Printf("[SQL] Open: open %s error: %s\n", driverName, err)
		stack[0] = 3
		return
	}
	// ping
	if err := db.Ping(); err != nil {
		log.Printf("[SQL] Open: ping error: %s\n", err)
		stack[0] = 4
		return
	}
	log.Println("[SQL] ✅ Open database success")
	// return db
	DB = db
	stack[0] = 0
}

// Query executes a query that returns rows, typically a SELECT
func Query(ctx context.Context, m api.Module, stack []uint64) {
	// query
	queryPtr := uint32(stack[0])
	querySize := uint32(stack[1])
	query, err := readBuffer(ctx, m, queryPtr, querySize)
	if err != nil {
		log.Printf("[SQL] Query: get query error: %s\n", err)
		stack[0] = 1
		return
	}
	// args
	var hasArgs bool
	var args []any
	argsPtr := uint32(stack[2])
	argsSize := uint32(stack[3])
	if argsPtr > 0 && argsSize > 0 {
		argsBuf, err := readBuffer(ctx, m, argsPtr, argsSize)
		if err != nil {
			log.Printf("[SQL] Query: get args error: %s\n", err)
			stack[0] = 2
			return
		}
		if err := json.Unmarshal(argsBuf, &args); err != nil {
			log.Printf("[SQL] Query: args unmarshal err: %s\n", err)
			stack[0] = 3
			return
		}
		hasArgs = len(args) > 0
	}
	// exec query
	var rows *sql.Rows
	if hasArgs {
		rows, err = DB.QueryContext(ctx, string(query), args...)
	} else {
		rows, err = DB.QueryContext(ctx, string(query))
	}
	if err != nil {
		log.Printf("[SQL] Query: execute(%s) error: %s\n", query, err)
		stack[0] = 4
		return
	}
	defer rows.Close()
	// result
	result, err := rows2maps(rows)
	if err != nil {
		log.Printf("[SQL] Query: rows2maps error: %s\n", err)
		stack[0] = 5
		return
	}
	resultBuf, err := json.Marshal(result)
	if err != nil {
		log.Printf("[SQL] Query: marshal error: %s\n", err)
		stack[0] = 6
		return
	}
	// allocate buffer and write to memory
	resultPtr := uint32(stack[4])
	resultSize := uint32(stack[5])
	if err := allocateBuffer(ctx, m, resultPtr, resultSize, resultBuf); err != nil {
		log.Printf("[SQL] Query: allocate buffer error: %s\n", err)
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
			log.Printf("[SQL] rows2map: scan error: %v\n", err)
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

// QueryRow query one row
func QueryRow(ctx context.Context, m api.Module, stack []uint64) {
	// query
	queryPtr := uint32(stack[0])
	querySize := uint32(stack[1])
	query, err := readBuffer(ctx, m, queryPtr, querySize)
	if err != nil {
		log.Printf("[SQL] QueryRow: get query error: %s\n", err)
		stack[0] = 1
		return
	}
	// args
	var hasArgs bool
	var args []any
	argsPtr := uint32(stack[2])
	argsSize := uint32(stack[3])
	if argsPtr > 0 && argsSize > 0 {
		argsBuf, err := readBuffer(ctx, m, argsPtr, argsSize)
		if err != nil {
			log.Printf("[SQL] QueryRow: get args error: %s\n", err)
			stack[0] = 2
			return
		}
		if err := json.Unmarshal(argsBuf, &args); err != nil {
			log.Printf("[SQL] QueryRow: args unmarshal err: %s\n", err)
			stack[0] = 3
			return
		}
		hasArgs = len(args) > 0
	}
	// exec query
	var rows *sql.Rows
	if hasArgs {
		rows, err = DB.QueryContext(ctx, string(query), args...)
	} else {
		rows, err = DB.QueryContext(ctx, string(query))
	}
	if err != nil {
		log.Printf("[SQL] QueryRow: execute(%s) error: %s\n", query, err)
		stack[0] = 4
		return
	}
	// result
	items, err := rows2maps(rows)
	if err != nil {
		log.Printf("[SQL] QueryRow: rows2maps error: %s\n", err)
		stack[0] = 5
		return
	}
	if len(items) > 0 {
		result := items[0]
		resultBuf, err := json.Marshal(result)
		if err != nil {
			log.Printf("[SQL] QueryRow: marshal error: %s\n", err)
			stack[0] = 6
			return
		}
		// allocate buffer and write to memory
		resultPtr := uint32(stack[4])
		resultSize := uint32(stack[5])
		if err := allocateBuffer(ctx, m, resultPtr, resultSize, resultBuf); err != nil {
			log.Printf("[SQL] Query: allocate buffer error: %s\n", err)
			stack[0] = 9
			return
		}
	}
	stack[0] = 0
}

// Exec execute sql
func Exec(ctx context.Context, m api.Module, stack []uint64) {
	// query
	queryPtr := uint32(stack[0])
	querySize := uint32(stack[1])
	query, err := readBuffer(ctx, m, queryPtr, querySize)
	if err != nil {
		log.Printf("[SQL] Exec: get query error: %s\n", err)
		stack[0] = 1
		return
	}
	// args
	var args []any
	argsPtr := uint32(stack[2])
	argsSize := uint32(stack[3])
	if argsPtr > 0 && argsSize > 0 {
		argsBuf, err := readBuffer(ctx, m, argsPtr, argsSize)
		if err != nil {
			log.Printf("[SQL] Exec: get args error: %s\n", err)
			stack[0] = 2
			return
		}
		if err := json.Unmarshal(argsBuf, &args); err != nil {
			log.Printf("[SQL] Exec: args unmarshal err: %s\n", err)
			stack[0] = 3
			return
		}
	}
	// exec query
	sqlResult, err := DB.ExecContext(ctx, string(query), args...)
	if err != nil {
		log.Printf("[SQL] Exec: execute(%s) error: %s\n", query, err)
		stack[0] = 4
		return
	}
	// result
	// LastInsertId returns the integer generated by the database
	// in response to a command. Typically this will be from an
	// "auto increment" column when inserting a new row. Not all
	// databases support this feature, and the syntax of such
	// statements varies.
	lastInsertID, _ := sqlResult.LastInsertId()
	// RowsAffected returns the number of rows affected by an
	// update, insert, or delete. Not every database or database
	// driver may support this.
	rowsAffected, _ := sqlResult.RowsAffected()
	result := map[string]int64{
		"last_insert_id": lastInsertID,
		"rows_affected":  rowsAffected,
	}
	resultBuf, err := json.Marshal(result)
	if err != nil {
		log.Printf("[SQL] Exec: marshal error: %s\n", err)
		stack[0] = 6
		return
	}
	// allocate buffer and write to memory
	resultPtr := uint32(stack[4])
	resultSize := uint32(stack[5])
	if err := allocateBuffer(ctx, m, resultPtr, resultSize, resultBuf); err != nil {
		log.Printf("[SQL] Exec: allocate buffer error: %s\n", err)
		stack[0] = 9
		return
	}
	stack[0] = 0
}

// Close closes the database and prevents new queries from starting
func Close(ctx context.Context, m api.Module, stack []uint64) {
	if err := DB.Close(); err != nil {
		log.Printf("[SQL] Close: %s\n", err)
		stack[0] = 1
		return
	}
	stack[0] = 0
}
