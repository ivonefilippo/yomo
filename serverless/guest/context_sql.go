package guest

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/yomorun/yomo/serverless"
)

// SQL is a database/sql wrapper
type GuestSQL struct{}

// Open opens a database specified by its database driver name and a
// driver-specific data source name, usually consisting of at least a
// database name and connection information
func (g *GuestSQL) Open(driverName string, dataSourceName string) error {
	driverNamePtr, driverNameSize := stringToPtrSize(driverName)
	if driverNameSize == 0 {
		return errors.New("driver name is empty")
	}
	dataSourceNamePtr, dataSourceNameSize := stringToPtrSize(dataSourceName)
	if dataSourceNameSize == 0 {
		return errors.New("data source name is empty")
	}
	// var db *sql.DB
	if ret := sqlOpen(driverNamePtr, driverNameSize, dataSourceNamePtr, dataSourceNameSize); ret != 0 {
		return fmt.Errorf("open sql error, driverName=%s, ret=%d", driverName, ret)
	}
	return nil
}

//
//export yomo_sql_open
func sqlOpen(
	driverNamePrt uintptr,
	driverNameSize uint32,
	dataSourceNamePtr uintptr,
	dataSourceNameSize uint32,
) uint32

// Query executes a query that returns rows, typically a SELECT
func (g *GuestSQL) Query(query string, args ...any) ([]map[string]any, error) {
	// args
	var argsPtr uintptr
	var argsSize uint32
	var hasArgs bool
	if len(args) > 0 {
		hasArgs = true
		argsBuf, err := json.Marshal(args)
		if err != nil {
			log.Printf("[GuestSQL] Query: get args error: %s\n", err)
			return nil, err
		}
		argsPtr, argsSize = bufferToPtrSize(argsBuf)
	}
	// query
	queryPtr, querySize := stringToPtrSize(query)
	if querySize == 0 {
		return nil, errors.New("query is empty")
	}
	// result
	var resultPtr *uint32
	var resultSize uint32
	var ret uint32
	if hasArgs {
		ret = sqlQuery(queryPtr, querySize, argsPtr, argsSize, &resultPtr, &resultSize)
	} else {
		ret = sqlQuery(queryPtr, querySize, 0, 0, &resultPtr, &resultSize)
	}
	if ret != 0 {
		err := fmt.Errorf("execute query error: %d", ret)
		log.Printf("[GuestSQL] Query: %s\n", err)
		return nil, err
	}
	resultBuf := readBufferFromMemory(resultPtr, resultSize)
	var result []map[string]any
	// json decode
	if err := json.Unmarshal(resultBuf, &result); err != nil {
		log.Printf("[GuestSQL] Query: result unmarshal error: %s\n", err)
		return nil, err
	}
	return result, nil
}

//export yomo_sql_query
func sqlQuery(
	queryPtr uintptr,
	querySize uint32,
	argsPtr uintptr,
	argsSize uint32,
	resultPtr **uint32,
	resultSize *uint32,
) uint32

// QueryRow executes a query that is expected to return at most one row
func (g *GuestSQL) QueryRow(query string, args ...any) (map[string]any, error) {
	// args
	var argsPtr uintptr
	var argsSize uint32
	var hasArgs bool
	if len(args) > 0 {
		hasArgs = true
		argsBuf, err := json.Marshal(args)
		if err != nil {
			log.Printf("[GuestSQL] QueryRow: get args error: %s\n", err)
			return nil, err
		}
		argsPtr, argsSize = bufferToPtrSize(argsBuf)
	}
	// query
	queryPtr, querySize := stringToPtrSize(query)
	if querySize == 0 {
		return nil, errors.New("query is empty")
	}
	// result
	var resultPtr *uint32
	var resultSize uint32
	var ret uint32
	if hasArgs {
		ret = sqlQueryRow(queryPtr, querySize, argsPtr, argsSize, &resultPtr, &resultSize)
	} else {
		ret = sqlQueryRow(queryPtr, querySize, 0, 0, &resultPtr, &resultSize)
	}
	if ret != 0 {
		err := fmt.Errorf("execute query error: %d", ret)
		log.Printf("[GuestSQL] QueryRow: %s\n", err)
		return nil, err
	}
	resultBuf := readBufferFromMemory(resultPtr, resultSize)
	var result map[string]any
	if len(resultBuf) == 0 {
		return nil, sql.ErrNoRows
	}
	// json decode
	if err := json.Unmarshal(resultBuf, &result); err != nil {
		log.Printf("[GuestSQL] QueryRow: result unmarshal error: %s\n", err)
		return nil, err
	}
	return result, nil
}

//export yomo_sql_query_row
func sqlQueryRow(
	queryPtr uintptr,
	querySize uint32,
	argsPtr uintptr,
	argsSize uint32,
	resultPtr **uint32,
	resultSize *uint32,
) uint32

// Exec executes a query without returning any rows
func (g *GuestSQL) Exec(query string, args ...any) (*serverless.SQLResult, error) {
	// args
	var argsPtr uintptr
	var argsSize uint32
	var hasArgs bool
	if len(args) > 0 {
		hasArgs = true
		argsBuf, err := json.Marshal(args)
		if err != nil {
			log.Printf("[GuestSQL] Exec: get args error: %s\n", err)
			return nil, err
		}
		argsPtr, argsSize = bufferToPtrSize(argsBuf)
	}
	// query
	queryPtr, querySize := stringToPtrSize(query)
	if querySize == 0 {
		return nil, errors.New("query is empty")
	}
	// result
	var resultPtr *uint32
	var resultSize uint32
	var ret uint32
	if hasArgs {
		ret = sqlExec(queryPtr, querySize, argsPtr, argsSize, &resultPtr, &resultSize)
	} else {
		ret = sqlExec(queryPtr, querySize, 0, 0, &resultPtr, &resultSize)
	}
	if ret != 0 {
		err := fmt.Errorf("execute query error: %d", ret)
		log.Printf("[GuestSQL] Exec: %s\n", err)
		return nil, err
	}
	resultBuf := readBufferFromMemory(resultPtr, resultSize)
	if len(resultBuf) == 0 {
		err := errors.New("execute query error: result is empty")
		log.Printf("[GuestSQL] Exec: %s\n", err)
		return nil, err
	}
	var result serverless.SQLResult
	// json decode
	if err := json.Unmarshal(resultBuf, &result); err != nil {
		log.Printf("[GuestSQL] Exec: result unmarshal error: %s\n", err)
		return nil, err
	}
	return &result, nil
}

//export yomo_sql_exec
func sqlExec(
	queryPtr uintptr,
	querySize uint32,
	argsPtr uintptr,
	argsSize uint32,
	resultPtr **uint32,
	resultSize *uint32,
) uint32

// Close closes the database and prevents new queries from starting
func (g *GuestSQL) Close() error {
	ret := sqlClose()
	if ret != 0 {
		err := fmt.Errorf("close sql error: %d", ret)
		log.Printf("[GuestSQL] Close: %s\n", err)
		return err
	}
	return nil
}

//export yomo_sql_close
func sqlClose() uint32
