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
		name      string
		in        io.Reader
		sensitive bool
		exp       string
		expErr    error
	}{
		{
			name:   "nil reader",
			expErr: errNilInput(),
		},
		{
			name: "non-array default behaviour",
			in:   sreader("{}"),
			exp:  "{}",
		},
		{
			name: "leading whitespace",
			in:   sreader("   \r\n{}"),
			exp:  "{}",
		},
		{
			name:   "just whitespace",
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
				tc.sensitive,
			}.run()
			assert.Equal(t, tc.exp, string(out.Bytes()), "expected output")
			assert.Equal(t, tc.expErr, err, "expected error")
		})
	}
}

func sreader(s string) io.Reader {
	return strings.NewReader(s)
}
