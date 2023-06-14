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

type HTTP interface {
	Get(url string) uint32
}
type SQL interface {
	Open(driverName string, dataSourceName string) error
	Query(query string) ([]map[string]any, error)
}
