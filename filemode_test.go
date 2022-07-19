package main

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileMode(t *testing.T) {

	cases := []struct {
		name     string
		files    []string
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
		// TODO BAD FILE
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out := bytes.NewBuffer(nil)
			err := filemode(tc.files, out, false)
			assert.Equal(t, tc.exp, string(out.Bytes()), "output")
			tc.checkErr(t, err)
		})
	}
}