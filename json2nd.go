package main

import (
	"fmt"
	"os"
)

func main() {

	if len(os.Args) > 1 {
		err := filemode(os.Args, os.Stdout, false)
		bail_if_err(err)
		return
	}

	err := processor{os.Stdin, os.Stdout, false}.run()
	bail_if_err(err)
}

func bail_if_err(e error) {
	if e == nil {
		return
	}

	fmt.Fprintln(os.Stderr, e)
}
