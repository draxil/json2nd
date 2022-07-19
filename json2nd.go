package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {

	tolerant, args := flags()

	if len(args) > 0 {
		err := filemode(args, os.Stdout, tolerant)
		bail_if_err(err)
		return
	}

	err := processor{os.Stdin, os.Stdout, tolerant}.run()
	bail_if_err(err)
}

func flags() (tolerant bool, args []string) {

	tolerantFlagValue := flag.Bool("expect-array", false, "check that whatever we're processing is an array, and fail if not.")

	flag.Parse()

	args = flag.Args()

	return *tolerantFlagValue, args
}

func bail_if_err(e error) {
	if e == nil {
		return
	}

	fmt.Fprintln(os.Stderr, e)
}
