package main

import (
	"encoding/json"
	"fmt"
	"io"
)

type processor struct {
	in       io.Reader
	out      io.Writer
	tolerant bool
}

// TODO: sensitive
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

	var msg json.RawMessage
	dec := json.NewDecoder(p.in)

	err := dec.Decode(&msg)
	if err != nil {
		return rawJSONErr(err)
	}

	var offset int
	for i, v := range msg {
		if !isSpace(v) {
			offset = i
			break
		}
		//TODO space only?
	}

	if msg[offset] != '[' {
		return p.handleNonArray(msg)
	}

	var bigArray []json.RawMessage
	err = json.Unmarshal(msg, &bigArray)
	if err != nil {
		return arrayJSONErr(err)
	}

	for _, v := range bigArray {
		// TODO partial? how does that work again?
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

func (p processor) handleNonArray(msg json.RawMessage) error {

	if p.tolerant {
		_, err := p.out.Write(msg)
		if err != nil {
			return err
		}
		return nil
	} else {
		return errNotArray()
	}

}

func errNilInput() error {
	return fmt.Errorf("nil input")
}

func errNotArray() error {
	// FUTURE: indicate what it was?
	return fmt.Errorf("expected structure to be an array")
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

// stolen from encoding/json
func isSpace(c byte) bool {
	return c <= ' ' && (c == ' ' || c == '\t' || c == '\r' || c == '\n')
}
