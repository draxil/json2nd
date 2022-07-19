package main

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProcessor(t *testing.T) {

	cases := []struct {
		name        string
		in          io.Reader
		expectArray bool
		exp         string
		expErr      error
	}{
		{
			name:   "nil reader",
			expErr: errNilInput(),
		},
		{
			name:        "non-array tolerant behaviour",
			expectArray: false,
			in:          sreader("{}"),
			exp:         "{}",
		},
		{
			name:        "non-array but expect one",
			expectArray: true,
			in:          sreader("{}"),
			expErr:      errNotArrayWas("object"),
		},
		{
			name:        "non-array + leading whitespace + tolerant",
			expectArray: false,
			in:          sreader("   \r\n{}"),
			exp:         "{}",
		},
		{
			name:        "just whitespace + tolerant",
			expectArray: false,
			in:          sreader("           "),
			expErr:      rawJSONErr(io.EOF),
		},
		{
			name:   "just whitespace not tolerant",
			in:     sreader("           "),
			expErr: rawJSONErr(io.EOF),
		},
		{
			name: "simple use-case",
			in:   sreader(`[{"a":1},{"b":2}]`),
			exp:  `{"a":1}` + "\n" + `{"b":2}` + "\n",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out := bytes.NewBuffer(nil)
			err := processor{
				tc.in,
				out,
				tc.expectArray,
			}.run()
			assert.Equal(t, tc.exp, string(out.Bytes()), "expected output")
			assert.Equal(t, tc.expErr, err, "expected error")
		})
	}
}

func TestGuessJsonType(t *testing.T) {

	cases := []struct {
		in  byte
		exp string
	}{
		{'{', "object"},
		{'"', "string"},
		{'5', "number"},
		{'1', "number"},
		{'0', "number"},
		{'9', "number"},
		{'x', ""},
		{'\'', ""},
		// we won't look for arrays:
		{'[', ""},
	}

	for _, tc := range cases {
		name := fmt.Sprintf("%c -> %s", tc.in, tc.exp)
		t.Run(name, func(t *testing.T) {
			get := guessJSONType(tc.in)
			assert.Equal(t, tc.exp, get)
		})
	}
}

func sreader(s string) io.Reader {
	return strings.NewReader(s)
}
