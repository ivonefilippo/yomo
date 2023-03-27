package frame

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGoawayFrameEncode(t *testing.T) {
	f := NewGoawayFrame("goaway")
	assert.Equal(t, "goaway", f.Message())
	assert.Equal(t, []byte{0x80 | byte(TagOfGoawayFrame), 0x8, 0x2, 0x6, 0x67, 0x6f, 0x61, 0x77, 0x61, 0x79}, f.Encode())
}

func TestGoawayFrameDecode(t *testing.T) {
	buf := []byte{0x80 | byte(TagOfGoawayFrame), 0x8, 0x2, 0x6, 0x67, 0x6f, 0x61, 0x77, 0x61, 0x79}
	f, err := DecodeToGoawayFrame(buf)
	assert.NoError(t, err)
	assert.Equal(t, []byte{0x80 | byte(TagOfGoawayFrame), 0x8, 0x2, 0x6, 0x67, 0x6f, 0x61, 0x77, 0x61, 0x79}, f.Encode())
}
