package main

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/draxil/json2nd/internal/json"
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
		{
			name:        "non-array tolerant behaviour",
			expectArray: false,
			in:          sreader("{}"),
			exp:         "{}\n",
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
			exp:         "{}\n",
		},
		{
			name:        "just whitespace + tolerant",
			expectArray: false,
			in:          sreader("           "),
			exp:         "",
			expErr:      errNoJSON(),
		},
		{
			name:        "just whitespace not tolerant",
			in:          sreader("           "),
			expectArray: true,
			expErr:      errNoJSON(),
		},
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
		{
			name:   "bad path",
			in:     sreader(`{}`),
			path:   "something",
			expErr: errBadPath("something"),
		},
		{
			name:   "bad path one down",
			in:     sreader(`{"something":{}}`),
			path:   "something.else",
			expErr: errBadPath("else"),
		},
		{
			name: "good simple path to string array",
			in:   sreader(`{"something":{"else":["one", "two"]}}`),
			path: "something.else",
			exp:  `"one"` + "\n" + `"two"` + "\n",
		},
		{
			name:   "broken path - 1",
			in:     sreader(`{"something":{}}`),
			path:   ".",
			expErr: errBlankPath(),
		},
		{
			name:   "broken path - 2",
			in:     sreader(`{"something":{}}`),
			path:   "something.",
			expErr: errBlankPath(),
		},
		{
			name:   "broken path - 2",
			in:     sreader(`{"something":{}}`),
			path:   "something..",
			expErr: errBlankPath(),
		},
		{
			name:   "broken path - 3",
			in:     sreader(`{"something":{}}`),
			path:   "..",
			expErr: errBlankPath(),
		},
		{
			name:   "broken path - 4",
			in:     sreader(`{"something":{}}`),
			path:   " ",
			expErr: errBadPath(" "),
		},
		{
			name:   "broken path - 5",
			in:     sreader(`{"something":{}}`),
			path:   "\u200B",
			expErr: errBadPath("\u200B"),
		},
		{
			name: "path leads to non-array",
			in:   sreader(`{"something":{}}`),
			path: "something",
			exp:  "{}\n",
		},
		{
			name:   "not JSON at all (looking for path)",
			in:     sreader("boo"),
			path:   "something",
			exp:    "",
			expErr: json.ErrScanNotObject{On: 'b'},
		},
		{
			name:   "path leads to non-JSON",
			in:     sreader(`{"something":boo}`),
			path:   "something",
			expErr: errPathLeadToBadValue('b'),
		},
		{
			name: "path leads to a number",
			in:   sreader(`{"something":12}`),
			path: "something",
			exp:  "12\n",
		},
		{
			name: "path leads to a space padded number",
			in:   sreader(`{"something":  12  }`),
			path: "something",
			exp:  "12\n",
		},
		{
			name: "path leads to a space padded float",
			in:   sreader(`{"something":  12.12  }`),
			path: "something",
			exp:  "12.12\n",
		},
		{
			name: "path leads to a space padded negative float",
			in:   sreader(`{"something":  -12.12  }`),
			path: "something",
			exp:  "-12.12\n",
		},
		{
			name: "path leads to a bool (true)",
			in:   sreader(`{"something":  true  }`),
			path: "something",
			exp:  "true\n",
		},
		{
			name: "path leads to a bool (false)",
			in:   sreader(`{"something":  false  }`),
			path: "something",
			exp:  "false\n",
		},
		{
			name: "path leads to a null",
			in:   sreader(`{"something":  null  }`),
			path: "something",
			exp:  "null\n",
		},
		{
			name:   "path leads to a bool (true), but with garbage afterwards",
			in:     sreader(`{"something":  truex  }`),
			path:   "something",
			exp:    "",
			expErr: json.ErrBadValue{"truex"},
		},
		// technically invalid JSON, but we're not a real parser ;)
		{
			name: "path leads to a bool (true), but with comma afterwards",
			in:   sreader(`{"something":  true,  }`),
			path: "something",
			exp:  "true\n",
		},

		{
			name:   "path goes through a null",
			in:     sreader(`{"something":  null }`),
			path:   "something.else",
			expErr: json.ErrScanNotObject{On: 'n'},
		},

		{
			name:   "path goes through a bool",
			in:     sreader(`{"something":  true }`),
			path:   "something.else",
			expErr: json.ErrScanNotObject{On: 't'},
		},

		{
			name:   "path goes through a negative number",
			in:     sreader(`{"something":  -129 }`),
			path:   "something.else",
			expErr: json.ErrScanNotObject{On: '-'},
		},
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
		{'-', "number"},
		{'x', ""},
		{'\'', ""},
		{'[', "array"},
	}

	for _, tc := range cases {
		name := fmt.Sprintf("%c -> %s", tc.in, tc.exp)
		t.Run(name, func(t *testing.T) {
			get := guessJSONType(tc.in)
			assert.Equal(t, tc.exp, get)
		})
	}
}

func TestErrPathLeadToBadValueMessage(t *testing.T) {

	cases := []struct {
		name     string
		clue     byte
		contains string
	}{
		{
			name:     "a number tells us it's a number type",
			clue:     '-',
			contains: "number",
		},
		{
			name:     "non-start character gives us a reasonable clue",
			clue:     'Z',
			contains: `Z`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := errPathLeadToBadValue(tc.clue)
			str := err.Error()
			assert.Contains(t, str, tc.contains)
		})
	}
}

func sreader(s string) io.Reader {
	return strings.NewReader(s)
}
