package main

import (
	"fmt"
	"io"
	"os"
)

func filemode(files []string, out io.Writer, expectArray bool, path string) error {

	for _, name := range files {
		f, err := os.Open(name)
		if err != nil {
			return fileOpenErr(name, err)
		}

		p := processor{f, out, expectArray, path}
		err = p.run()
		if err != nil {
			return fileProcessErr(name, err)
		}
	}
	return nil
}

func fileOpenErr(file string, e error) error {
	return fmt.Errorf("could not open %s: %w", file, e)
}

func fileProcessErr(file string, e error) error {
	return fmt.Errorf("could not process %s: %w", file, e)
}

func fmWriteErr(e error) error {
	return fmt.Errorf("could not write to output: %w", e)
}
