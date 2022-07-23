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

// TODO: dive down mode
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

	if c != '[' {
		panic("not implemented on internal json branch")
		//return p.handleNonArray(pr, c)
	}

	// TODO BACK IN FUNC

	return p.handleArray(js)
	// var bigArray []json.RawMessage
	// err = dec.Decode(&bigArray)
	// if err != nil {
	// 	return arrayJSONErr(err)
	// }

	// return p.handleArray(bigArray)
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
		// TODO:
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

		panic("NDY")
		//return p.handleNonArray(scan, clue)
	}
	return p.handlePathNodes(nodes, scan)
}

// 	var nextObj map[string]json.RawMessage
// 	err := json.Unmarshal(target, &nextObj)
// 	if err != nil {
// 		// TODO: hard to cover as would be caught earlier,
// 		// but maybe when we change JSON method?
// 		return fmt.Errorf("could not decode path node %s: %w", next, err)
// 	}

// 	return p.handlePathNodes(nodes, nextObj)
// }
func (p processor) handleArray(js *json.JSON) error {

	// TODO more in json

	// shift the cursor from the start of the array:
	js.MoveOff()
	for {
		nc, err := js.Next()
		if err != nil {
			// TODO WHEN IN FUNC BETTER ERRS
			return arrayJSONErr(err)
		}
		if nc == ',' {
			js.MoveOff()
			continue
		}
		if nc == ']' {
			break
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
		js.MoveOff()
	}
	return nil
}

func (p processor) handleNonArray(r io.Reader, clue byte) error {

	if !p.expectArray {
		_, err := io.Copy(p.out, r)
		if err != nil {
			return err
		}
		return nil
	} else {
		return errNotArray(clue)
	}

}

func guessJSONType(clue byte) string {

	switch clue {
	case '{':
		return "object"
	case '"':
		return "string"
	}

	if clue >= '0' && clue <= '9' {
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
