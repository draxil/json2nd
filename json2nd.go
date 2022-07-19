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

	tolerantFlagValue := flag.Bool("tolerant", false, "Be tolerant: if the structure found is not an array, output it anyway without translation")

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
