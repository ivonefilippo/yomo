package wazero

import (
	"context"
	"errors"

	"github.com/tetratelabs/wazero/api"
)

// allocateBuffer allocates memory and writes the data to the memory
func allocateBuffer(
	ctx context.Context,
	m api.Module,
	bufPtr uint32,
	bufSize uint32,
	buf []byte,
) error {
	bufLen := len(buf)
	memResults, err := m.ExportedFunction("yomo_alloc").Call(ctx, uint64(bufLen))
	if err != nil {
		return err
	}
	allocPtr := uint32(memResults[0])
	if !m.Memory().WriteUint32Le(bufPtr, allocPtr) {
		return errors.New("memory write `bufPtr` error")
	}
	if !m.Memory().WriteUint32Le(bufSize, uint32(bufLen)) {
		return errors.New("memory write `bufSize` error")
	}
	if !m.Memory().Write(allocPtr, buf) {
		return errors.New("memory write `buf` error")
	}
	return nil
}
