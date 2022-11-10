package main

import (
	"flag"
	"fmt"
	"os"
	"runtime/debug"

	"github.com/draxil/json2nd/internal/options"
)

var version = ""

func main() {

	oh, err := options.New(os.Args[1:])
	if err != nil {
		if err == flag.ErrHelp {
			os.Exit(0)
		}
		bail(err)
	}

	args := oh.Args()
	opts := oh.Options

	if opts.JustPrintVersion {
		justPrintVersion()
	}

	// TODO: JUST PASS OPTIONS
	if len(args) > 0 {
		err := filemode(args, os.Stdout, opts.ExpectArray, opts.Path)
		bailIfError(err)
		return
	}

	err = processor{os.Stdin, os.Stdout, opts.ExpectArray, opts.Path, true}.run()
	bailIfError(err)
}

func bailIfError(e error) {
	if e == nil {
		return
	}

	bail(e)
}

func bail(e error) {
	fmt.Fprintln(os.Stderr, e)
	os.Exit(1)
}

func justPrintVersion() {
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
