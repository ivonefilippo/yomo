package serverless

// Context sfn handler context
type Context interface {
	// Data incoming data
	Data() []byte
	// Tag incoming tag
	Tag() uint32
	// Write write data to zipper
	Write(tag uint32, data []byte) error
	// HTTP http package
	HTTP() HTTP
	// SQL sql package
	SQL() SQL
}

// HTTP http interface
type HTTP interface {
	Get(url string) uint32
}

// SQL sql database interface
type SQL interface {
	Open(driverName string, dataSourceName string) error
	Query(query string, args ...any) ([]map[string]any, error)
	QueryRow(query string, args ...any) (map[string]any, error)
	Exec(query string, args ...any) (*SQLResult, error)
	Close() error
}

// SQLResult sql result
type SQLResult struct {
	LastInsertID int64 `json:"last_insert_id"`
	RowsAffected int64 `json:"rows_affected"`
}
