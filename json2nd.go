package main

import (
	"fmt"
	"os"
)

func main() {
	err := processor{os.Stdin, os.Stdout, false}.run()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}
