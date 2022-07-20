package peeky

import (
	"bytes"
	"fmt"
	"io"
)

const chunkSize = 8

// peeky.NonWSReader wraps a reader such that we can look for the first non-ws character without
// loosing anything
type NonWSReader struct {
	buf        []byte
	underlying io.Reader
}

func NewNonWSReader(w io.Reader) *NonWSReader {
	return &NonWSReader{nil, w}
}

func (n *NonWSReader) Peek() (byte, error) {
	var res byte
	buf := bytes.NewBuffer(make([]byte, 0, chunkSize))
	idx := 0

	for res == 0 {
		var chunk []byte = make([]byte, chunkSize)

		read, err := n.underlying.Read(chunk)
		if err != nil && err != io.EOF {
			return 0, err
		}

		for _, v := range chunk {
			if isSpace(v) {
				continue
			}

			res = v
			break
		}

		_, werr := buf.Write(chunk[:read])
		if werr != nil {
			return 0, fmt.Errorf("could not write to buffer")
		}

		if err == io.EOF {
			break
		}

		idx += read
	}

	n.buf = buf.Bytes()

	return res, nil
}

func (n *NonWSReader) Read(b []byte) (int, error) {
	if n.buf != nil {
		moved := copy(b, n.buf)
		if moved < len(n.buf) {
			n.buf = n.buf[moved:]
		} else {
			n.buf = nil
		}
		return moved, nil
	}
	return n.underlying.Read(b)
}

// stolen from encoding/json
func isSpace(c byte) bool {
	return c <= ' ' && (c == ' ' || c == '\t' || c == '\r' || c == '\n')
}
