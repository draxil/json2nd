package json

import (
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNextSimple(t *testing.T) {
	r := strings.NewReader("   [")
	j := New(r)

	c, err := j.Next()
	assert.Equal(t, byte('['), c, "start")
	assert.NoError(t, err)
}

func sread(s string) io.Reader {
	return strings.NewReader(s)
}

func TestNext(t *testing.T) {

	cases := []struct {
		name   string
		reader io.Reader
		exp    byte
		expErr error
	}{
		{
			name:   "one char",
			reader: sread("["),
			exp:    '[',
			expErr: nil,
		},
		{
			name:   "just ws",
			reader: sread(" "),
			exp:    0,
			expErr: io.EOF,
		},
		{
			name:   "empty reader",
			reader: sread(""),
			exp:    0,
			expErr: io.EOF,
		},
		{
			name:   "all our WS",
			reader: sread("\n\r\t '"),
			exp:    '\'',
			expErr: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			get, err := New(tc.reader).Next()
			assert.Equal(t, tc.exp, get, "value")
			assert.Equal(t, tc.expErr, err, "error value")
		})
	}
}

func TestWriteToSimple(t *testing.T) {
	out := strings.Builder{}
	in := sread("    \n [1,2,3,4] ")
	j := New(in)

	get, err := j.Next()
	assert.Equal(t, byte('['), get, "start")
	assert.NoError(t, err, "no error on Next() call")

	n, err := j.WriteTo(&out)
	assert.NoError(t, err, "no error on WriteTo")
	assert.Equal(t, 9, n, "n")
	assert.Equal(t, "[1,2,3,4]", out.String(), "output")
}

func TestWriteToTinyChunk(t *testing.T) {
	out := strings.Builder{}
	in := sread("    \n [1,2,3,4] ")
	j := New(in)
	j.chunkSize = 2

	get, err := j.Next()
	assert.Equal(t, byte('['), get, "start")
	assert.NoError(t, err, "no error on Next() call")

	n, err := j.WriteTo(&out)
	assert.NoError(t, err, "no error on WriteTo")
	assert.Equal(t, 9, n, "n")
	assert.Equal(t, "[1,2,3,4]", out.String(), "output")
}

// WRITETO:
// NEST ARRAYS
// STRINGS WITH ARRAY CHARS
// STRINGS WITH ESCAPE QUOTES
// OBJECT
// NESTED OBJECT
// STRING
