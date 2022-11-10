package options

import (
	"flag"
)

const (
	path     = "path"
	exparray = "expect-array"
	version  = "version"
)

// New create an option handler that will parse the options from command line args
func New(args []string) (Handler, error) {
	var h Handler
	var o Options

	h.FlagSet = flag.NewFlagSet("json2nd", flag.ContinueOnError)

	h.StringVar(
		&o.Path,
		path,
		"",
		"path to get to the JSON value you want to extract, e.g key1.key2",
	)
	h.BoolVar(
		&o.ExpectArray,
		exparray,
		false,
		"check that whatever we're processing is an array, and fail if not",
	)
	h.BoolVar(
		&o.JustPrintVersion,
		version,
		false,
		"print the version description for this tool and exit",
	)

	err := h.Parse(args)
	h.Options = o

	return h, err
}

// TODO actual usage etc
// TODO: this doesn't work because we want the options to be a separate thing:

type Handler struct {
	*flag.FlagSet
	Options Options
}

type Options struct {
	ExpectArray      bool
	JustPrintVersion bool
	Path             string
	Args             []string
}
