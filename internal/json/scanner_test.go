package json

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStateScanSimpleObject(t *testing.T) {
	s := NewScanState('{')
	assert.False(t, s.key)

	_, err := s.scan([]byte{'"'}, 0, 1)
	assert.NoError(t, err)
	assert.True(t, s.inStr)
	assert.True(t, s.key)

	_, err = s.scan([]byte{'\\'}, 0, 1)
	assert.NoError(t, err)
	assert.True(t, s.inStr)
	assert.True(t, s.key)

	_, err = s.scan([]byte{'"'}, 0, 1)
	assert.NoError(t, err)
	assert.True(t, s.inStr)
	assert.True(t, s.key)

	_, err = s.scan([]byte{'"'}, 0, 1)
	assert.NoError(t, err)
	assert.False(t, s.inStr)
	assert.False(t, s.key)
	assert.True(t, s.open, "still open after key")

	_, err = s.scan([]byte(` : `), 0, 3)
	assert.NoError(t, err)
	assert.False(t, s.inStr)
	assert.False(t, s.key)

	_, err = s.scan([]byte(`"y`), 0, 2)
	assert.NoError(t, err)
	assert.True(t, s.inStr)
	assert.False(t, s.key)

	_, err = s.scan([]byte(`ellow"`), 0, 6)
	assert.NoError(t, err)
	assert.False(t, s.inStr)
	assert.True(t, s.open, "still open after value")

	_, err = s.scan([]byte{'}'}, 0, 1)
	assert.NoError(t, err)
	assert.False(t, s.open, "closed by the closer")
}

func TestScanSubObjectDoesNotClose(t *testing.T) {
	s := NewScanState('{')
	buf := []byte(`"x":{"subobject":21}`)
	_, err := s.scan(buf, 0, len(buf))
	assert.True(t, s.open)
	assert.NoError(t, err)

	_, err = s.scan([]byte{'}'}, 0, 1)
	assert.False(t, s.open)
	assert.NoError(t, err)
}

func TestScanCloseBeforeEnd(t *testing.T) {
	s := NewScanState('{')
	buf := []byte(`"x":"y"}      foo`)
	pos, _ := s.scan(buf, 0, len(buf))
	assert.False(t, s.open)
	assert.Equal(t, buf[pos], byte('}'), "cursor ends where we expect")

}

func TestScanStringDoesNotCloseObject(t *testing.T) {
	s := NewScanState('{')

	buf := []byte(`"}"`)
	_, err := s.scan(buf, 0, len(buf))
	assert.NoError(t, err)
	assert.True(t, s.open)

	_, err = s.scan([]byte{'}'}, 0, 1)
	assert.NoError(t, err)
	assert.False(t, s.open)
}

func TestEscapedSubStringDoesNotClose(t *testing.T) {
	s := NewScanState('"')
	buf := []byte(`x\"`)
	_, err := s.scan(buf, 0, len(buf))
	assert.NoError(t, err)
	assert.True(t, s.open)

	_, err = s.scan([]byte{'"'}, 0, 1)
	assert.NoError(t, err)
	assert.False(t, s.open)
}

func TestScanOnChar(t *testing.T) {
	s := NewScanState('Z')
	buf := []byte(`x\"`)
	_, err := s.scan(buf, 0, len(buf))
	assert.Equal(t, ErrBadJSONValue{'Z'}, err)
}

func TestScanForSimple(t *testing.T) {
	s := NewScanState('{')
	s.seekFor("z")
	buf := []byte(`"x":99,"y":102,`)

	_, err := s.scan(buf, 0, len(buf))
	assert.False(t, s.seekFound, "not found yet")
	assert.NoError(t, err)

	buf = []byte(`"z":199`)
	_, err = s.scan(buf, 0, len(buf))
	assert.NoError(t, err)
	assert.True(t, s.seekFound, "found")

}

func TestScanForSimpleWithNestedTrap(t *testing.T) {
	s := NewScanState('{')
	s.seekFor("target")
	assert.False(t, s.seekFound, "not found yet")
	buf := []byte(`"x":99,"p":{"target":12},`)

	_, err := s.scan(buf, 0, len(buf))
	assert.NoError(t, err)
	assert.False(t, s.seekFound, "not found yet")

	buf = []byte(`"target":199`)
	pos, err := s.scan(buf, 0, len(buf))
	assert.NoError(t, err)
	assert.True(t, s.seekFound, "found")
	assert.Equal(t, `":199`, string(buf[pos:]), "stopped in the correct place")
}

func TestFutileScan(t *testing.T) {
	s := NewScanState('{')
	s.seekFor("target")
	assert.False(t, s.seekFound, "not found yet")

	buf := []byte(`"x":99,"p":{"target":12},"z":[{"target":12}]}`)
	_, err := s.scan(buf, 0, len(buf))
	assert.False(t, s.seekFound, "not found yet")
	assert.False(t, s.open)
	assert.NoError(t, err)
}
