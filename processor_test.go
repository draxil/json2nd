package main

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProcessor(t *testing.T) {

	cases := []struct {
		name     string
		in       io.Reader
		tolerant bool
		exp      string
		expErr   error
	}{
		{
			name:   "nil reader",
			expErr: errNilInput(),
		},
		{
			name:     "non-array tolerant behaviour",
			tolerant: true,
			in:       sreader("{}"),
			exp:      "{}",
		},
		{
			name:     "non-array default behaviour",
			tolerant: false,
			in:       sreader("{}"),
			expErr:   errNotArray(),
		},
		{
			name:     "non-array + leading whitespace + tolerant",
			tolerant: true,
			in:       sreader("   \r\n{}"),
			exp:      "{}",
		},
		{
			name:     "just whitespace + tolerant",
			tolerant: true,
			in:       sreader("           "),
			expErr:   rawJSONErr(io.EOF),
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
				tc.tolerant,
			}.run()
			assert.Equal(t, tc.exp, string(out.Bytes()), "expected output")
			assert.Equal(t, tc.expErr, err, "expected error")
		})
	}
}

func sreader(s string) io.Reader {
	return strings.NewReader(s)
}
