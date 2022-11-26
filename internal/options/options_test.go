package options

import (
	"flag"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOptParse(t *testing.T) {

	cases := []struct {
		name     string
		in       []string
		exp      Set
		checkErr func(t *testing.T, e error)
	}{
		{
			name: "no options",
			in:   []string{},
			exp:  Set{},
			checkErr: func(t *testing.T, e error) {
				assert.NoError(t, e)
			},
		},
		{
			name: "version",
			in:   []string{"-version"},
			exp: Set{
				JustPrintVersion: true,
			},
			checkErr: func(t *testing.T, e error) {
				assert.NoError(t, e)
			},
		},
		{
			name: "path",
			in:   []string{"-path", "xyz.boo"},
			exp: Set{
				Path: "xyz.boo",
			},
			checkErr: func(t *testing.T, e error) {
				assert.NoError(t, e)
			},
		},
		{
			name: "expect array",
			in:   []string{"-expect-array"},
			exp: Set{
				ExpectArray: true,
			},
			checkErr: func(t *testing.T, e error) {
				assert.NoError(t, e)
			},
		},
		{
			name: "bad arg",
			in:   []string{"-size"},
			checkErr: func(t *testing.T, e error) {
				assert.Error(t, e)
			},
		},
		{
			name: "preserve array",
			in:   []string{"-preserve-array"},
			exp: Set{
				PreserveArray: true,
			},
			checkErr: func(t *testing.T, e error) {
				assert.NoError(t, e)
			},
		},
		{
			name: "preserve array + expect array",
			in:   []string{"-preserve-array", "-expect-array"},
			exp:  Set{},
			checkErr: func(t *testing.T, e error) {
				is := assert.Error(t, e)
				if !is {
					return
				}
				assert.Contains(t, e.Error(), "options conflict", "error message")
			},
		},
		{
			name: "help",
			in:   []string{"-help"},
			checkErr: func(t *testing.T, e error) {
				assert.Equal(t, e, flag.ErrHelp)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			get, err := New(tc.in)
			assert.Equal(t, tc.exp, get.Options, "result")
			if tc.checkErr != nil {
				tc.checkErr(t, err)
			}
		})
	}
}

//TODO method to get the option description we need for error messages
