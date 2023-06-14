package guest

import "unsafe"

type GuestHTTP struct{}

//export yomo_http_get
//go:linkname httpGet
func httpGet(ptr uintptr, size uint32) uint32

func (g *GuestHTTP) Get(url string) uint32 {
	reqURL := []byte(url)
	ptr := uintptr(unsafe.Pointer(&reqURL[0]))
	length := uint32(len(reqURL))
	return httpGet(ptr, length)
}
