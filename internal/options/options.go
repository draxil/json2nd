package options

import (
	"flag"
	"fmt"
)

const (
	OptPath          = "path"
	OptExpectArray   = "expect-array"
	OptVersion       = "version"
	OptPreserveArray = "preserve-array"
)

// New create an option handler that will parse the options from command line args
func New(args []string) (Handler, error) {
	var h Handler
	var o Set

	h.FlagSet = flag.NewFlagSet("json2nd", flag.ContinueOnError)

	h.StringVar(
		&o.Path,
		OptPath,
		"",
		"path to get to the JSON value you want to extract, e.g key1.key2",
	)
	h.BoolVar(
		&o.ExpectArray,
		OptExpectArray,
		false,
		"check that whatever we're processing is an array, and fail if not",
	)
	h.BoolVar(
		&o.JustPrintVersion,
		OptVersion,
		false,
		"print the version description for this tool and exit",
	)
	h.BoolVar(
		&o.PreserveArray,
		OptPreserveArray,
		false,
		"instead of turning the top-level array into NDJSON preserve the array, useful for JSON streams",
	)

	err := h.Parse(args)

	if o.PreserveArray && o.ExpectArray {
		return h, fmt.Errorf("options conflict, -%s does not work alongside -%s", OptPreserveArray, OptExpectArray)
	}

	h.Options = o

	return h, err
}

type Handler struct {
	*flag.FlagSet
	Options Set
}

type Set struct {
	ExpectArray      bool
	PreserveArray    bool
	JustPrintVersion bool
	Path             string
	Args             []string
}
