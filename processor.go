package main

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/draxil/json2nd/internal/json"
	"github.com/draxil/json2nd/internal/options"
)

type processor struct {
	in       io.Reader
	out      io.Writer
	options  options.Set
	buffered bool
}

// TODO: detect where not an object more tidily in path mode
// TODO: no error on bad path mode?
// TODO: builds?
// TODO: how should default be for formatting
// TODO: opt for un-formatted (OR THE OTHER WAY AROUND?)
// TODO: path to value?

func (p processor) run() error {

	if p.in == nil {
		return errNilInput()
	}

	js := json.New(p.in)
	if p.options.Path != "" {
		return p.handlePath(js)
	}

	c, err := js.Next()
	if err != nil && err != io.EOF {
		return peekErr(err)
	}
	if err == io.EOF {
		return errNoJSON()
	}

	if c != '[' || p.options.PreserveArray {
		return p.handleNonArray(js, c, true)
	}

	return p.handleArray(js)
}

func (p processor) handlePath(scan *json.JSON) error {
	nodes := strings.Split(p.options.Path, ".")
	return p.handlePathNodes(nodes, scan)
}

func (p processor) handlePathNodes(nodes []string, scan *json.JSON) error {

	// shouldn't be possible? But never say never.
	if len(nodes) == 0 {
		return fmt.Errorf("novel error 1: please report")
	}

	next, nodes := nodes[0], nodes[1:]
	if next == "" {
		return errBlankPath()
	}

	found, err := scan.ScanForKeyValue(next)
	if err != nil {
		return err
	}
	if !found {
		return errBadPath(next)
	}

	if len(nodes) == 0 {

		clue, err := scan.Next()
		if err != nil {
			// TODO:
			return err
		}

		if clue == '[' {
			return p.handleArray(scan)
		}

		return p.handleNonArray(scan, clue, false)
	}
	return p.handlePathNodes(nodes, scan)
}

func (p processor) prepOut() (w io.Writer, finishOut func() error) {
	if p.buffered {
		bw := bufio.NewWriter(p.out)
		return bw, bw.Flush
	}
	return p.out, func() error { return nil }
}

func (p processor) handleArray(js *json.JSON) error {

	// shift the cursor from the start of the array:
	js.MoveOff()

	out, finishOut := p.prepOut()

	arrayIDX := 0
	for {
		c, err := js.Next()
		if err != nil {
			return arrayNextError(arrayIDX, err)
		}

		if !json.SaneValueStart(c) {
			return errBadArrayValueStart(c, arrayIDX)
		}

		n, err := js.WriteCurrentTo(out, true)

		if err != nil {
			return arrayJSONErr(err)
		}
		if n > 0 {
			_, err := out.Write([]byte("\n"))
			if err != nil {
				return arrayJSONErr(err)
			}
		}
		if n == 0 {
			break
		}

		c, err = js.Next()
		if err != nil {
			return arrayNextError(arrayIDX, err)
		}

		if c == ']' {
			break
		}
		if c == ',' {
			js.MoveOff()
		} else {
			return fmt.Errorf("unexpected character: %c", c)
		}

		arrayIDX++
	}

	return finishOut()
}

func (p processor) handleNonArray(j *json.JSON, clue byte, topLevel bool) error {
	out, finishOut := p.prepOut()

	if p.options.ExpectArray {
		return errNotArrayWas(guessJSONType(j.Peek()))
	}
	if !json.SaneValueStart(clue) {
		return errPathLeadToBadValue(clue, p.options.Path)
	}

	for {
		n, err := j.WriteCurrentTo(out, true)
		if err != nil {
			if err == io.EOF {
				return errNonArrayEOF(guessJSONType(clue))
			}
			return err
		}
		if n > 0 {
			_, err := out.Write([]byte("\n"))
			if err != nil {
				return err
			}
		}

		// we don't continue if we scanned down to this level
		if !topLevel {
			break
		}

		// okay now we're in some kind of JSON stream like
		// NDJSON, so look for the next thing
		clue, err := j.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if clue == '[' && !p.options.PreserveArray {
			return errArrayInStream()
		}

	}

	return finishOut()
}

func guessJSONType(clue byte) string {

	switch clue {
	case '{':
		return "object"
	case '"':
		return "string"
	case '[':
		return "array"
	case 't':
		return "boolean"
	case 'f':
		return "boolean"
	case 'n':
		return "null"
	}

	if (clue >= '0' && clue <= '9') || clue == '-' {
		return "number"
	}

	return ""
}

func errNilInput() error {
	return fmt.Errorf("nil input")
}

func errNotArrayWas(t string) error {
	return fmt.Errorf("expected structure to be an array but found: %s", t)
}

func errNonArrayEOF(t string) error {
	if t == "" {
		return fmt.Errorf("JSON data ended prematurely")
	}
	return fmt.Errorf("file ended before the end of value (%s)", t)
}

func errBadPath(chunk string) error {
	return fmt.Errorf("path node did not exist: %s", chunk)
}

func errBlankPath() error {
	return fmt.Errorf("bad blank path node, did you have a double dot?")
}
func errPathLeadToBadValue(start byte, path string) error {
	t := guessJSONType(start)

	if t != "" {
		return fmt.Errorf("path (%s) lead to %s not an object", path, t)
	}

	return fmt.Errorf("path (%s) lead to bad value start: %c", path, start)
}

func errBadArrayValueStart(start byte, index int) error {
	return fmt.Errorf("at array index %d found something which doesn't look like a JSON value, starts with: %c", index, start)
}

func arrayNextError(index int, e error) error {
	if e == io.EOF {
		return errArrayEOF(index)
	}
	// TODO: process this?
	return e
}

func errArrayEOF(index int) error {
	return fmt.Errorf("at array index %d we ran out of data", index)
}

func errNoJSON() error {
	return fmt.Errorf("no JSON data found")
}

func arrayJSONErr(e error) error {
	return fmt.Errorf("array JSON decode error: %w", e)
}

func peekErr(e error) error {
	return fmt.Errorf("error looking for JSON: %w", e)
}

func errArrayInStream() error {
	return fmt.Errorf("Found an array in what seemed to be a JSON stream. Consider using -%s", options.OptPreserveArray)
}
