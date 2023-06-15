// guest wasm application programming interface for guest module
package guest

import (
	"unsafe"
)

var (
	// ReadBuf is a buffer used to read data from the host.
	ReadBuf = make([]byte, ReadBufSize)
	// ReadBufPtr is a pointer to ReadBuf.
	ReadBufPtr = uintptr(unsafe.Pointer(&ReadBuf[0]))
	// ReadBufSize is the size of ReadBuf
	ReadBufSize = uint32(2048)
)

// GetBytes returns a byte slice of the given size
func GetBytes(fn func(ptr uintptr, size uint32) (len uint32)) (result []byte) {
	size := fn(ReadBufPtr, ReadBufSize)
	if size == 0 {
		return
	}
	if size > 0 && size <= ReadBufSize {
		// copy to avoid passing a mutable buffer
		result = make([]byte, size)
		copy(result, ReadBuf)
		return
	}
	// Otherwise, allocate a new buffer
	buf := make([]byte, size)
	ptr := uintptr(unsafe.Pointer(&buf[0]))
	_ = fn(ptr, size)
	return buf
}

//export yomo_alloc
func alloc(size uint32) uintptr {
	buf := make([]byte, size)
	ptr := &buf[0]
	return uintptr(unsafe.Pointer(ptr))
}

// stringToPtrSize converts a string to a pointer and its size.
func stringToPtrSize(s string) (uintptr, uint32) {
	if s == "" {
		return 0, 0
	}
	buf := []byte(s)
	ptr := &buf[0]
	unsafePtr := uintptr(unsafe.Pointer(ptr))
	return unsafePtr, uint32(len(buf))
}

// bufferPtrSize returns the memory position and size of the buffer
func bufferToPtrSize(buff []byte) (uintptr, uint32) {
	ptr := &buff[0]
	unsafePtr := uintptr(unsafe.Pointer(ptr))
	return unsafePtr, uint32(len(buff))
}

func packPtrAndSize(ptr uint32, size uint32) uint64 {
	return uint64(ptr)<<32 | uint64(size)
}

func unpackPtrAndSize(ptrAndSize uint64) (uint32, uint32) {
	return uint32(ptrAndSize >> 32), uint32(ptrAndSize)
}

// readBufferFromMemory returns a buffer
func readBufferFromMemory(bufferPosition *uint32, length uint32) []byte {
	buf := make([]byte, length)
	ptr := uintptr(unsafe.Pointer(bufferPosition))
	for i := 0; i < int(length); i++ {
		s := *(*int32)(unsafe.Pointer(ptr + uintptr(i)))
		buf[i] = byte(s)
	}
	return buf
}

// copyBufferToMemory returns a single value (a kind of pair with position and size)
func copyBufferToMemory(buffer []byte) uint64 {
	bufferPtr := &buffer[0]
	unsafePtr := uintptr(unsafe.Pointer(bufferPtr))

	ptr := uint32(unsafePtr)
	size := uint32(len(buffer))

	return packPtrAndSize(ptr, size)
}
