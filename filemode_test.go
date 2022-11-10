package main

import (
	"bytes"
	"errors"
	"testing"

	"github.com/draxil/json2nd/internal/options"
	"github.com/stretchr/testify/assert"
)

func TestFileMode(t *testing.T) {

	cases := []struct {
		name     string
		files    []string
		opts     options.Set
		exp      string
		checkErr func(t *testing.T, e error)
	}{
		{
			name:  "file does not exist",
			files: []string{"./fictional.does.not.exist"},
			checkErr: func(t *testing.T, e error) {
				iserr := assert.Error(t, e, "there is an error")
				if !iserr {
					return
				}
				assert.Contains(t, e.Error(), "could not open", "context")
				inner := errors.Unwrap(e)
				assert.Error(t, inner, "there is an inner error")
			},
		},
		{
			name:  "file does exist",
			files: []string{"./testdata/1.json"},
			checkErr: func(t *testing.T, e error) {
				assert.NoError(t, e)
			},
			exp: `{"one":1}` + "\n",
		},
		{
			name:  "two files exist",
			files: []string{"./testdata/1.json", "./testdata/2.json"},
			checkErr: func(t *testing.T, e error) {
				assert.NoError(t, e)
			},
			exp: `{"one":1}` + "\n" + `{"two":2}` + "\n" + `{"three":3}` + "\n",
		},
		{
			name:  "with a path",
			files: []string{"./testdata/simpleobj.json"},
			checkErr: func(t *testing.T, e error) {
				assert.NoError(t, e)
			},
			opts: options.Set{
				Path: "x",
			},
			exp: `1` + "\n" + `2` + "\n" + `4` + "\n",
		},
		// TODO BAD FILE
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out := bytes.NewBuffer(nil)
			err := filemode(tc.files, out, tc.opts)
			assert.Equal(t, tc.exp, out.String(), "output")
			tc.checkErr(t, err)
		})
	}
}
