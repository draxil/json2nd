package main

import (
	"flag"
	"fmt"
	"os"
	"runtime/debug"
)

var version = ""

func main() {

	tolerant, path, args := flags()

	if len(args) > 0 {
		err := filemode(args, os.Stdout, tolerant, path)
		bail_if_err(err)
		return
	}

	err := processor{os.Stdin, os.Stdout, tolerant, path, true}.run()
	bail_if_err(err)
}

func flags() (tolerant bool, path string, args []string) {

	tolerantFlagValue := flag.Bool("expect-array", false, "check that whatever we're processing is an array, and fail if not.")

	pathValue := flag.String("path", "", "path to get to the JSON value you want to extract e.g key1.key2")

	showVersion := flag.Bool("version", false, "print the version description for this tool and exit")

	flag.Parse()

	args = flag.Args()

	if *showVersion {
		if version != "" {
			fmt.Println(version)
		} else {
			info, ok := debug.ReadBuildInfo()
			if ok {
				fmt.Println(info.Main.Version)
			} else {
				fmt.Println("don't know")
			}
		}

		os.Exit(0)
	}

	return *tolerantFlagValue, *pathValue, args
}

func bail_if_err(e error) {
	if e == nil {
		return
	}

	fmt.Fprintln(os.Stderr, e)
	os.Exit(1)
}
