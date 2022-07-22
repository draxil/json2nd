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

	n, err := j.WriteTo(&out, true)
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

	n, err := j.WriteTo(&out, true)
	assert.NoError(t, err, "no error on WriteTo")
	assert.Equal(t, 9, n, "n")
	assert.Equal(t, "[1,2,3,4]", out.String(), "output")
}

func TestWriteToTinyChunkNoDelims(t *testing.T) {
	out := strings.Builder{}
	in := sread("    \n [1,2,3,4] ")
	j := New(in)
	j.chunkSize = 2

	get, err := j.Next()
	assert.Equal(t, byte('['), get, "start")
	assert.NoError(t, err, "no error on Next() call")

	n, err := j.WriteTo(&out, false)
	assert.NoError(t, err, "no error on WriteTo")
	assert.Equal(t, 7, n, "n")
	assert.Equal(t, "1,2,3,4", out.String(), "output")
}

func TestWriteToByteChunkNoDelims(t *testing.T) {
	out := strings.Builder{}
	in := sread("    \n [1,2,3,4] ")
	j := New(in)
	j.chunkSize = 1

	get, err := j.Next()
	assert.Equal(t, byte('['), get, "start")
	assert.NoError(t, err, "no error on Next() call")

	n, err := j.WriteTo(&out, false)
	assert.NoError(t, err, "no error on WriteTo")
	assert.Equal(t, 7, n, "n")
	assert.Equal(t, "1,2,3,4", out.String(), "output")
}

func TestWriteTo(t *testing.T) {

	cases := []struct {
		name    string
		in      io.Reader
		delims  bool
		exp     string
		expClue byte
	}{
		{
			name:    "nested array",
			in:      sread("  [[1],[2]] "),
			delims:  true,
			exp:     "[[1],[2]]",
			expClue: '[',
		},
		{
			name:    "object array",
			in:      sread("[{},{}] "),
			delims:  true,
			exp:     "[{},{}]",
			expClue: '[',
		},
		{
			name:    "no delims",
			in:      sread("[{},{}] "),
			delims:  false,
			exp:     "{},{}",
			expClue: '[',
		},
		{
			name:    "object array + keys",
			in:      sread("  [{\"x\":{}}] "),
			delims:  false,
			exp:     "{\"x\":{}}",
			expClue: '[',
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			j := New(tc.in)
			// lets have an awkward chunk for these
			j.chunkSize = 3
			clue, err := j.Next()
			assert.Equal(t, tc.expClue, clue, "start")
			assert.NoError(t, err, "no error on Next() call")

			out := strings.Builder{}
			n, err := j.WriteTo(&out, tc.delims)
			assert.NoError(t, err, "no error on WriteTo")
			assert.Equal(t, tc.exp, out.String(), "output")
			assert.Equal(t, len(tc.exp), n, "n")
		})
	}
}

// WRITETO:
// NEST ARRAYS
// STRINGS WITH ARRAY CHARS
// STRINGS WITH ESCAPE QUOTES
// OBJECT
// NESTED OBJECT
// STRING
