package json

import (
	"fmt"
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
	in := sread("    \n [1,2,3,4]   ")
	j := New(in)

	get, err := j.Next()
	assert.Equal(t, byte('['), get, "start")
	assert.NoError(t, err, "no error on Next() call")

	n, err := j.WriteCurrentTo(&out, true)
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

	n, err := j.WriteCurrentTo(&out, true)
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

	n, err := j.WriteCurrentTo(&out, false)
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

	n, err := j.WriteCurrentTo(&out, false)
	assert.NoError(t, err, "no error on WriteTo")
	assert.Equal(t, 7, n, "n")
	assert.Equal(t, "1,2,3,4", out.String(), "output")
}

func TestCurrentWriteTo(t *testing.T) {

	cases := []struct {
		name     string
		in       io.Reader
		delims   bool
		exp      string
		expClue  byte
		checkErr func(t *testing.T, e error)
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
		{
			name:    "target: object",
			in:      sread("  {\"x\":\"foo\"} "),
			delims:  true,
			exp:     "{\"x\":\"foo\"}",
			expClue: '{',
		},
		{
			name:    "target: object - no delims",
			in:      sread("  {\"x\":\"foo\"} "),
			delims:  false,
			exp:     "\"x\":\"foo\"",
			expClue: '{',
		},
		{
			name:    "closer appears inside a string",
			in:      sread(`["]"]`),
			delims:  false,
			exp:     `"]"`,
			expClue: '[',
		},
		{
			name:    "escaped quotes in strings",
			in:      sread(`["]\""]`),
			delims:  false,
			exp:     `"]\""`,
			expClue: '[',
		},
		{
			name:    "target a string",
			in:      sread(`   "1, 2, 3, 4" `),
			delims:  true,
			exp:     `"1, 2, 3, 4"`,
			expClue: '"',
		},
		{
			name:    "target is a number",
			in:      sread(`12`),
			delims:  false,
			exp:     "12",
			expClue: '1',
		},
		{
			name:    "target is a number - with trailing stuff",
			in:      sread(`12,`),
			delims:  false,
			exp:     "12",
			expClue: '1',
		},
		{
			name:    "target is a floating point number",
			in:      sread(`12.12`),
			delims:  false,
			exp:     "12.12",
			expClue: '1',
		},
		{
			name:    "target is true",
			in:      sread(`true`),
			delims:  false,
			exp:     "true",
			expClue: 't',
		},
		{
			name:    "target is false",
			in:      sread(`false`),
			delims:  false,
			exp:     "false",
			expClue: 'f',
		},
		{
			name:    "target is null",
			in:      sread(`null`),
			delims:  false,
			exp:     "null",
			expClue: 'n',
		},
		{
			name:    "target in notation: 1e-005",
			in:      sread("1e-005,"),
			delims:  false,
			exp:     "1e-005",
			expClue: '1',
		},
		{
			name:    "target in notation: 1.0E+2",
			in:      sread("1.0E+2,"),
			delims:  false,
			exp:     "1.0E+2",
			expClue: '1',
		},
		{
			name:    "target is a negative number",
			in:      sread("-12"),
			delims:  false,
			exp:     "-12",
			expClue: '-',
		},
		{
			name:    "target is a negative float",
			in:      sread("-12.01"),
			delims:  false,
			exp:     "-12.01",
			expClue: '-',
		},
		{
			name:    "show effect of delims on an awkward number",
			in:      sread(" -12.01-210, "),
			delims:  true,
			exp:     "-12.01-210",
			expClue: '-',
		},
		{
			name:    "simple newline infested array (+delims)",
			in:      sread("\n[\n1\n,\n2\n,\n3\n\n,\n4\n]\n"),
			delims:  true,
			exp:     "[1,2,3,4]",
			expClue: '[',
		},
		{
			name:    "simple newline infested array (-delims)",
			in:      sread("\n[\n1\n,\n2\n,\n3\n\n,\n4\n]\n"),
			delims:  false,
			exp:     "1,2,3,4",
			expClue: '[',
		},
		{
			name:    "simple newline infested object (+delims)",
			in:      sread("\n{\"one\"\n:\n1\n,\"two\"\n:\n2\n,\n\"three\"\n:\n3\n\n,\n\"four\"\n:4\n}\n"),
			delims:  true,
			exp:     "{\"one\":1,\"two\":2,\"three\":3,\"four\":4}",
			expClue: '{',
		},
		{
			name:    "simple newline infested object (-delims)",
			in:      sread("\n{\"one\"\n:\n1\n,\"two\"\n:\n2\n,\n\"three\"\n:\n3\n\n,\n\"four\"\n:4\n}\n"),
			delims:  false,
			exp:     "\"one\":1,\"two\":2,\"three\":3,\"four\":4",
			expClue: '{',
		},
		// not valid JSON but we tolerate it currently:
		{
			name:    "newline within a string",
			in:      sread(`"` + "foo\nbar" + `"`),
			delims:  false,
			exp:     "foo\nbar",
			expClue: '"',
		},
		{
			name:    "array does not close, no tailing comma",
			in:      sread("[1,3"),
			delims:  false,
			exp:     "1,3",
			expClue: '[',
			checkErr: func(t *testing.T, e error) {
				assert.Equal(t, io.EOF, e)
			},
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
			n, err := j.WriteCurrentTo(&out, tc.delims)
			if tc.checkErr == nil {
				assert.NoError(t, err, "no error on WriteTo")
			} else {
				tc.checkErr(t, err)
			}
			assert.Equal(t, tc.exp, out.String(), "output")
			assert.Equal(t, len(tc.exp), n, "n")
		})
	}
}

func TestScanForKey(t *testing.T) {

	cases := []struct {
		name     string
		reader   io.Reader
		key      string
		expErr   error
		expFound bool
	}{
		{
			name:     "next error",
			reader:   sread(""),
			key:      "x",
			expErr:   io.EOF,
			expFound: false,
		},
		{
			name:     "not object error",
			reader:   sread(" ["),
			key:      "x",
			expErr:   ErrScanNotObject{'['},
			expFound: false,
		},
		{
			name:     "key not found",
			reader:   sread("{}"),
			key:      "x",
			expFound: false,
		},
		{
			name:     "ignore nested object with wrong key",
			reader:   sread(`{"xyz":{"x":12}}`),
			key:      "x",
			expFound: false,
		},
		{
			name:     "don't find value as key",
			reader:   sread(`{"v":"x"}`),
			key:      "x",
			expFound: false,
		},
		{
			name:     "do find key (simple)",
			reader:   sread(`{"x":"v"}`),
			key:      "x",
			expErr:   nil,
			expFound: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			j := New(tc.reader)
			found, err := j.ScanForKey(tc.key)
			assert.Equal(t, tc.expFound, found, "found")
			assert.Equal(t, tc.expErr, err, "err")
		})
	}
}

func TestScanForKeyValueSimple(t *testing.T) {
	r := sread(`{"x":"v"}`)
	j := New(r)

	found, err := j.ScanForKeyValue("x")
	assert.NoError(t, err)
	assert.True(t, found, "found")

	b := strings.Builder{}
	_, err = j.WriteCurrentTo(&b, true)

	assert.NoError(t, err)
	assert.Equal(t, `"v"`, b.String(), "value")
}

func TestSaneValueStart(t *testing.T) {

	cases := []struct {
		in  byte
		exp bool
	}{
		{'{', true},
		{'}', false},
		{'[', true},
		{']', false},
		{'b', false},
		{'"', true},
		{'n', true},
		{'x', false},
		{'f', true},
		// JSON numbers can't actually start with a 0:
		{'0', false},
		{'1', true},
		{'2', true},
		{'3', true},
		{'4', true},
		{'5', true},
		{'6', true},
		{'7', true},
		{'8', true},
		{'9', true},
		{'@', false},
		{'\'', false},
		// for negative numbers:
		{'-', true},
	}

	for _, tc := range cases {
		name := fmt.Sprintf("%c", tc.in)
		t.Run(name, func(t *testing.T) {
			get := SaneValueStart(tc.in)
			assert.Equal(t, tc.exp, get)
		})
	}
}
