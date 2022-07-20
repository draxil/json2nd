package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {

	tolerant, path, args := flags()

	if len(args) > 0 {
		err := filemode(args, os.Stdout, tolerant, path)
		bail_if_err(err)
		return
	}

	err := processor{os.Stdin, os.Stdout, tolerant, path}.run()
	bail_if_err(err)
}

func flags() (tolerant bool, path string, args []string) {

	tolerantFlagValue := flag.Bool("expect-array", false, "check that whatever we're processing is an array, and fail if not.")

	pathValue := flag.String("path", "", "path to get to the JSON value you want to extract e.g key1.key2")

	flag.Parse()

	args = flag.Args()

	return *tolerantFlagValue, *pathValue, args
}

func bail_if_err(e error) {
	if e == nil {
		return
	}

	fmt.Fprintln(os.Stderr, e)
}
