package peeky

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWsPeekSimple(t *testing.T) {
	b := []byte("       [")
	r := bytes.NewReader(b)
	pr := NewNonWSReader(r)
	get, err := pr.Peek()
	assert.NoError(t, err)
	assert.Equal(t, byte('['), get, "value")
}

func TestWsPeekBeyondChunk(t *testing.T) {
	b := []byte("        [")
	r := bytes.NewReader(b)
	pr := NewNonWSReader(r)
	get, err := pr.Peek()
	assert.NoError(t, err)
	assert.Equal(t, byte('['), get, "value")
}

func TestWsPeekJustWS(t *testing.T) {
	b := []byte("       ")
	r := bytes.NewReader(b)
	pr := NewNonWSReader(r)
	get, err := pr.Peek()
	assert.NoError(t, err)
	assert.Equal(t, byte(0), get, "value")
}

func TestWsPeekShort(t *testing.T) {
	b := []byte("")
	r := bytes.NewReader(b)
	pr := NewNonWSReader(r)
	get, err := pr.Peek()
	assert.NoError(t, err)
	assert.Equal(t, byte(0), get, "value")
}

func TestWsPeekWithNL(t *testing.T) {
	b := []byte("      \n[")
	r := bytes.NewReader(b)
	pr := NewNonWSReader(r)
	get, err := pr.Peek()
	assert.NoError(t, err)
	assert.Equal(t, byte('['), get, "value")
}

func TestReadNoPeek(t *testing.T) {
	b := []byte("       [")
	r := bytes.NewReader(b)
	pr := NewNonWSReader(r)
	get, err := io.ReadAll(pr)
	assert.NoError(t, err)
	assert.Equal(t, []byte("       ["), get, "read value")
}

func TestReadWithPeek(t *testing.T) {
	b := []byte("       [")
	r := bytes.NewReader(b)

	pr := NewNonWSReader(r)
	get, err := pr.Peek()
	assert.NoError(t, err)
	assert.Equal(t, byte('['), get, "value")

	getr, err := io.ReadAll(pr)
	assert.NoError(t, err)
	assert.Equal(t, []byte("       ["), getr, "read value")
}

func TestReadLessThanBuf(t *testing.T) {
	b := []byte("       [")
	r := bytes.NewReader(b)

	pr := NewNonWSReader(r)
	get, err := pr.Peek()
	assert.NoError(t, err)
	assert.Equal(t, byte('['), get, "value")

	r1 := make([]byte, 4)
	n, err := pr.Read(r1)
	assert.NoError(t, err, "no error")
	assert.Equal(t, 4, n, "copied 4")
	assert.Equal(t, "    ", string(r1), "read 1 value")

	r2 := make([]byte, 4)
	n, err = pr.Read(r2)
	assert.NoError(t, err, "no error on read 2")
	assert.Equal(t, 4, n, "copied 4 on read 2")
	assert.Equal(t, "   [", string(r2), "read 2 value")
}

func TestReadWithPeekShort(t *testing.T) {
	b := []byte("")
	r := bytes.NewReader(b)

	pr := NewNonWSReader(r)
	get, err := pr.Peek()
	assert.NoError(t, err)
	assert.Equal(t, byte(0), get)

	getr, err := io.ReadAll(pr)
	assert.NoError(t, err)
	assert.Equal(t, []byte(""), getr, "read value")
}
