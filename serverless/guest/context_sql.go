package guest

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
)

// SQL is a database/sql wrapper.
type GuestSQL struct{}

// Open opens a database specified by its database driver name and a
// driver-specific data source name, usually consisting of at least a
// database name and connection information.
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

func (g *GuestSQL) Query(query string) ([]map[string]any, error) {
	queryPtr, querySize := stringToPtrSize(query)
	if querySize == 0 {
		return nil, errors.New("query is empty")
	}
	var resultPtr *uint32
	var resultSize uint32
	ret := sqlQuery(queryPtr, querySize, &resultPtr, &resultSize)
	if ret != 0 {
		log.Printf("[GuestSQL] Query error, ret=%d\n", ret)
		return nil, errors.New("query error")
	}
	resultBuf := readBufferFromMemory(resultPtr, resultSize)
	log.Printf("[GuestSQL] Query resultBuf=%s\n", resultBuf)
	var result []map[string]any
	// json decode
	err := json.Unmarshal(resultBuf, &result)
	if err != nil {
		log.Printf("[GuestSQL] Unmarshal error, err=%v\n", err)
		return nil, err
	}
	return result, nil
}

//export yomo_sql_query
func sqlQuery(queryPtr uintptr, querySize uint32, resultPtr **uint32, resultSize *uint32) uint32
