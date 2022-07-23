package json

// TODO: MASTER OFFSET FOR ERRORS?

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
	if err == io.EOF {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return true, nil
}

func (j *JSON) MoveOff() {
	j.idx++
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

func closerFor(b byte) byte {
	switch b {
	case '[':
		return ']'
	case '{':
		return '}'
	case '"':
		return '"'
	}
	return 0
}
func (j *JSON) ScanForKeyValue(k string) (bool, error) {
	found, err := j.ScanForKey(k)
	if err != nil {
		return false, err
	}
	if !found {
		return false, nil
	}

	j.MoveOff()
	c, err := j.Next()
	if err != nil {
		return false, err
	}
	if c != ':' {
		return false, fmt.Errorf("expected ':' found %c", c)
	}

	j.MoveOff()
	_, err = j.Next()
	if err != nil {
		return false, err
	}
	// should be on a value now

	return true, nil
}
func (j *JSON) ScanForKey(k string) (bool, error) {
	start, err := j.Next()
	if err != nil {
		return false, err
	}

	if start != '{' {
		return false, ErrScanNotObject{start}
	}

	// kick cursor into the object
	j.MoveOff()

	scanner := NewScanState(start)
	scanner.seekFor(k)

	for {
		var err error

		j.idx, err = scanner.scan(j.buf, j.idx, j.bytes)
		if err != nil {
			// TODO: context?
			return false, err
		}

		if scanner.seekFound {
			return true, nil
		}

		if !scanner.open {
			return false, nil
		}

		if j.idx == j.bytes {
			more, err := j.data()
			if err != nil {
				// TODO: context?
				return false, err
			}
			if !more {
				break
			}
		}
	}

	return false, nil
}

type ErrScanNotObject struct {
	on byte
}

func (ErrScanNotObject) Error() string {
	return "can't scan for key, not on an object"
}

func (j *JSON) WriteCurrentTo(w io.Writer, includeDeliminators bool) (int, error) {
	// TODO: do we care about numbers, bools, null being json values? If not do we really care about strings?

	// because Next() leaves us resting on the start
	start := j.buf[j.idx]

	if start == 0 {
		return 0, errNoObject()
	}

	scanner := NewScanState(start)

	alreadyWritten := 0
	wn := 0
	end := 0
	var err error

	if includeDeliminators {
		_, err := w.Write([]byte{start})
		if err != nil {
			return 0, err
		}
	}

	j.MoveOff()

	for {
		end, err = scanner.scan(j.buf, j.idx, j.bytes)
		if err != nil {
			return 0, err
		}

		// in this case we're on the delim:
		if includeDeliminators && !scanner.open {
			end++
			alreadyWritten++
		}

		wn, err = w.Write(j.buf[j.idx:end])
		if err != nil {
			return 0, err
		}

		j.idx = end
		alreadyWritten += wn

		if scanner.open {
			more, err := j.data()
			if err != nil {
				// TODO: context?
				return 0, err
			}
			if !more {
				break
			}
		} else {

			break
		}
	}

	return alreadyWritten, nil
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
