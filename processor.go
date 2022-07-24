package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/draxil/json2nd/internal/json"
)

type processor struct {
	in          io.Reader
	out         io.Writer
	expectArray bool
	path        string
}

// TODO: detect where not an object more tidily in path mode
// TODO: prove buf
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
	if p.path != "" {
		return p.handlePath(js)
	}

	c, err := js.Next()
	if err != nil && err != io.EOF {
		return peekErr(err)
	}
	if err == io.EOF {
		return errNoJSON()
	}

	if c != '[' {
		return p.handleNonArray(js, c)
	}

	return p.handleArray(js)
}

func (p processor) handlePath(scan *json.JSON) error {
	nodes := strings.Split(p.path, ".")
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

		return p.handleNonArray(scan, clue)
	}
	return p.handlePathNodes(nodes, scan)
}

func (p processor) handleArray(js *json.JSON) error {

	// TODO more in json

	// shift the cursor from the start of the array:
	js.MoveOff()

	for {
		c, err := js.Next()
		if err != nil {
			return err
		}

		n, err := js.WriteCurrentTo(p.out, true)

		if err != nil {
			return arrayJSONErr(err)
		}
		if n > 0 {
			_, err := p.out.Write([]byte("\n"))
			if err != nil {
				return arrayJSONErr(err)
			}
		}
		if n == 0 {
			break
		}

		c, err = js.Next()
		if err != nil {
			return err
		}

		if c == ']' {
			return nil
		}
		if c == ',' {
			js.MoveOff()
		} else {
			return fmt.Errorf("unexpected character: %c", c)
		}

	}
	return nil
}

func (p processor) handleNonArray(j *json.JSON, clue byte) error {

	if p.expectArray {
		return errNotArrayWas(guessJSONType(j.Peek()))
	}
	if !json.SaneValueStart(clue) {
		return errPathLeadToBadValue(clue)
	}

	n, err := j.WriteCurrentTo(p.out, true)
	if err != nil {
		return err
	}
	if n > 0 {
		_, err := p.out.Write([]byte("\n"))
		if err != nil {
			return err
		}
	}
	return nil
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

func errNotArray(clue byte) error {
	t := guessJSONType(clue)

	if t != "" {
		return errNotArrayWas(t)
	}

	// FUTURE: indicate what it was?
	return fmt.Errorf("expected structure to be an array")
}

func errNotArrayWas(t string) error {
	return fmt.Errorf("expected structure to be an array but found: %s", t)
}

func errBadPath(chunk string) error {
	return fmt.Errorf("path node did not exist: %s", chunk)
}

func errBlankPath() error {
	return fmt.Errorf("bad blank path node, did you have a double dot?")
}
func errPathLeadToBadValue(start byte) error {
	t := guessJSONType(start)

	// TODO: reference the path
	if t != "" {
		return fmt.Errorf("path lead to %s not an object", t)
	}

	return fmt.Errorf("path lead to bad value start: %c", start)
}

func errNoJSON() error {
	return fmt.Errorf("no JSON data found")
}

func rawJSONErr(e error) error {
	return fmt.Errorf("raw JSON decode error: %w", e)
}

func arrayJSONErr(e error) error {
	return fmt.Errorf("array JSON decode error: %w", e)
}

func readErr(e error) error {
	return fmt.Errorf("read error: %w", e)
}

func seekErr(e error) error {
	return fmt.Errorf("seek error: %w", e)
}

func writeErr(e error) error {
	return fmt.Errorf("write error: %w", e)
}

func peekErr(e error) error {
	return fmt.Errorf("error looking for JSON: %w", e)
}

// stolen from encoding/json
func isSpace(c byte) bool {
	return c <= ' ' && (c == ' ' || c == '\t' || c == '\r' || c == '\n')
}
