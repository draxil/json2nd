package main

import (
	"encoding/json"
	"fmt"
	"github.com/draxil/json2nd/internal/peeky"
	"io"
)

type processor struct {
	in          io.Reader
	out         io.Writer
	expectArray bool
}

// TODO: reconsider tolerant in view of http://ndjson.org/
// TODO: dive down mode
// TODO: peeky reader idea
// TODO: prove buf
// TODO: filename mode
// TODO: builds?
// TODO: default to formatted
// TODO: opt for unformatted (OR THE OTHER WAY AROUND?)
// TODO: path to value?

func (p processor) run() error {

	if p.in == nil {
		return errNilInput()
	}

	pr := peeky.NewNonWSReader(p.in)

	dec := json.NewDecoder(pr)

	c, err := pr.Peek()
	if err != nil && err != io.EOF {
		return peekErr(err)
	}

	if c != '[' {
		return p.handleNonArray(pr, c)
	}

	var bigArray []json.RawMessage
	err = dec.Decode(&bigArray)
	if err != nil {
		return arrayJSONErr(err)
	}

	for _, v := range bigArray {
		_, err := p.out.Write(v)
		if err != nil {
			return writeErr(err)
		}
		_, err = p.out.Write([]byte("\n"))
		if err != nil {
			return writeErr(err)
		}
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
