package json

import (
	"fmt"
	"io"
)

type JSON struct {
	r         io.Reader
	buf       []byte
	idx       int
	bytes     int
	chunkSize int
}

func New(r io.Reader) *JSON {
	const defaultChunkSize = 4096
	return &JSON{r, nil, -1, 0, defaultChunkSize}
}

func (j *JSON) data() (bool, error) {

	if j.buf == nil {
		j.buf = make([]byte, j.chunkSize)
	}

	if j.idx >= 0 && j.idx < j.bytes {
		return true, nil
	}

	j.idx = 0

	var err error
	j.bytes, err = j.r.Read(j.buf)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (j *JSON) Next() (c byte, e error) {
	is, err := j.data()

	if err != nil && err != io.EOF {
		return 0, err
	}
	if !is {
		return 0, io.EOF
	}

	for j.idx < j.bytes {
		if !isSpace(j.buf[j.idx]) {
			return j.buf[j.idx], nil
		}
		j.idx++
	}

	return j.Next()
}

// TODO: proper implementation
func closerFor(b byte) byte {
	return ']'
}

func (j *JSON) WriteTo(w io.Writer) (int, error) {

	// because Next() leaves us resting on the start
	startIdx := j.idx
	start := j.buf[j.idx]
	j.idx++

	if start == 0 {
		return 0, errNoObject()
	}

	closer := closerFor(start)
	if closer == 0 {
		return 0, ErrInternal{fmt.Errorf("no closer for: %c", start)}
	}

	end := -1
	closerBalance := 0
	alreadyWritten := 0

	for {
		for ; j.idx < len(j.buf); j.idx++ {
			b := j.buf[j.idx]
			if b == closer {
				if closerBalance > 0 {
					closerBalance--
					continue
				}

				end = j.idx
				break
			}
			if b == start {
				closerBalance++
			}
		}

		if end != -1 {
			break
		}

		n, err := w.Write(j.buf[startIdx:j.bytes])
		if err != nil {
			return 0, err
		}
		alreadyWritten += n
		startIdx = 0

		// TODO: will chuck data
		more, err := j.data()
		if err != nil {
			// TODO: context?
			return 0, err
		}
		if !more {
			break
		}
	}

	if end == -1 {
		// TODO: clarity?
		return 0, io.EOF
	}

	os := j.buf[startIdx : end+1]
	n, err := w.Write(os)
	if err != nil {
		return alreadyWritten, err
	}
	return n + alreadyWritten, nil
}

func errNoObject() error {
	// TODO: internal error system?
	return ErrInternal{fmt.Errorf("no current object")}
}

// TODO: unwrappable
type ErrInternal struct {
	inner error
}

func (e ErrInternal) Error() string {
	return fmt.Sprintf("internal error: %v", e.inner)
}

func isSpace(c byte) bool {
	return c <= ' ' && (c == ' ' || c == '\t' || c == '\r' || c == '\n')
}
