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
		path        string
		expectArray bool
		exp         string
		errChecker  func(*testing.T, error)
		expErr      error
	}{
		{
			name:   "nil reader",
			expErr: errNilInput(),
		},
		// {
		// 	name:        "non-array tolerant behaviour",
		// 	expectArray: false,
		// 	in:          sreader("{}"),
		// 	exp:         "{}",
		// },
		// {
		// 	name:        "non-array but expect one",
		// 	expectArray: true,
		// 	in:          sreader("{}"),
		// 	expErr:      errNotArrayWas("object"),
		// },
		// {
		// 	name:        "non-array + leading whitespace + tolerant",
		// 	expectArray: false,
		// 	in:          sreader("   \r\n{}"),
		// 	// TODO: future, might strip or?
		// 	exp: "   \r\n{}",
		// },
		// {
		// 	// FUTURE: is this desirable?
		// 	name:        "just whitespace + tolerant",
		// 	expectArray: false,
		// 	in:          sreader("           "),
		// 	exp:         "           ",
		// 	expErr:      nil,
		// },
		// {
		// 	name:        "just whitespace not tolerant",
		// 	in:          sreader("           "),
		// 	expectArray: true,
		// 	expErr:      errNotArray(' '),
		// },
		{
			name: "simple use-case",
			in:   sreader(`[{"a":1},{"b":2}]`),
			exp:  `{"a":1}` + "\n" + `{"b":2}` + "\n",
		},

		{
			name: "nested array",
			in:   sreader(`[[{"a":1},{"b":2}], [2]]`),
			exp:  `[{"a":1},{"b":2}]` + "\n" + `[2]` + "\n",
		},
		// {
		// 	name:   "bad path",
		// 	in:     sreader(`{}`),
		// 	path:   "something",
		// 	expErr: errBadPath("something"),
		// },
		// {
		// 	name:   "bad path one down",
		// 	in:     sreader(`{"something":{}}`),
		// 	path:   "something.else",
		// 	expErr: errBadPath("else"),
		// },
		// {
		// 	name: "good simple path to string array",
		// 	in:   sreader(`{"something":{"else":["one", "two"]}}`),
		// 	path: "something.else",
		// 	exp:  `"one"` + "\n" + `"two"` + "\n",
		// },
		// {
		// 	name:   "broken path - 1",
		// 	in:     sreader(`{"something":{}}`),
		// 	path:   ".",
		// 	expErr: errBlankPath(),
		// },
		// {
		// 	name:   "broken path - 2",
		// 	in:     sreader(`{"something":{}}`),
		// 	path:   "something.",
		// 	expErr: errBlankPath(),
		// },
		// {
		// 	name:   "broken path - 2",
		// 	in:     sreader(`{"something":{}}`),
		// 	path:   "something..",
		// 	expErr: errBlankPath(),
		// },
		// {
		// 	name:   "broken path - 3",
		// 	in:     sreader(`{"something":{}}`),
		// 	path:   "..",
		// 	expErr: errBlankPath(),
		// },
		// {
		// 	name:   "broken path - 4",
		// 	in:     sreader(`{"something":{}}`),
		// 	path:   " ",
		// 	expErr: errBadPath(" "),
		// },
		// {
		// 	name:   "broken path - 5",
		// 	in:     sreader(`{"something":{}}`),
		// 	path:   "\u200B",
		// 	expErr: errBadPath("\u200B"),
		// },
		// {
		// 	name: "path leads to non-array",
		// 	in:   sreader(`{"something":{}}`),
		// 	path: "something",
		// 	exp:  "{}", // TODO: NL
		// },
		// {
		// 	name: "path leads to non-JSON",
		// 	in:   sreader(`{"something":boo}`),
		// 	path: "something",
		// 	errChecker: func(t *testing.T, e error) {
		// 		is := assert.Error(t, e)
		// 		if !is {
		// 			return
		// 		}
		// 		assert.Contains(t, e.Error(), "raw JSON decode error:", "looks like a JSON error")
		// 	},
		//	},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out := bytes.NewBuffer(nil)
			err := processor{
				tc.in,
				out,
				tc.expectArray,
				tc.path,
			}.run()
			assert.Equal(t, tc.exp, string(out.Bytes()), "expected output")
			if tc.errChecker != nil {
				tc.errChecker(t, err)
			} else {
				assert.Equal(t, tc.expErr, err, "expected error")
			}
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
