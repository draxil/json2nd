package json

// TODO: MASTER OFFSET FOR ERRORS?
// TODO: unicode considerations?
// TODO: BETTER ERROR HANDLING

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

func (j *JSON) Peek() byte {
	if j.buf == nil {
		return 0
	}
	if !(j.idx < len(j.buf)) {
		return 0
	}
	return j.buf[j.idx]
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
	On byte
}

func (ErrScanNotObject) Error() string {
	return "can't scan for key, not on an object"
}

func (j *JSON) WriteCurrentTo(w io.Writer, includeDeliminators bool) (int, error) {
	// TODO: note sure all delims schenarios are covered with no delim types

	// because Next() leaves us resting on the start
	start := j.buf[j.idx]

	if start == 0 {
		return 0, errNoObject()
	}

	if (start >= '0' && start <= '9') || start == '-' {
		return j.writeCurrentNumber(w)
	}
	if start == 't' || start == 'f' || start == 'n' {
		return j.writeKeywordValue(w)
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

		if j.idx < j.bytes && j.buf[j.idx] == '\n' && end > j.idx {
			end--
		}

		wn, err = w.Write(j.buf[j.idx:end])
		if err != nil {
			return 0, err
		}

		if j.idx < j.bytes && j.buf[j.idx] == '\n' {
			end++
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
				return alreadyWritten, io.EOF
			}
		} else {

			break
		}
	}

	return alreadyWritten, nil
}

// NOTE: Would allow some bad values such 1.....4 or 1e+-2134 but remember
// JSON validation isn't really our jam.
func (j *JSON) writeCurrentNumber(w io.Writer) (int, error) {
	end := false
	written := 0
	start := 0

	for !end {
		start = j.idx

		for ; j.idx < j.bytes; j.idx++ {
			c := j.buf[j.idx]
			if !((c >= '0' && c <= '9') ||
				c == '.' ||
				c == 'e' ||
				c == 'E' ||
				c == '-' ||
				c == '+') {
				end = true
				break
			}
		}

		n, err := w.Write(j.buf[start:j.idx])
		if err != nil {
			return 0, err
		}

		written += n
		if !end && j.idx == j.bytes {
			more, err := j.data()
			if err != nil {
				return 0, err
			}
			if !more {
				break
			}
		}

	}

	return written, nil
}

func (j *JSON) writeKeywordValue(w io.Writer) (int, error) {
	end := false
	stored := 0
	written := 0
	var keyword [5]byte

	for !end {
		for ; j.idx < j.bytes && stored < 5; j.idx++ {
			c := j.buf[j.idx]
			if !(c >= 'a' && c <= 'z') {
				end = true
				break
			} else {
				keyword[stored] = c
				stored++
			}
		}

		if !end && j.idx == j.bytes {
			more, err := j.data()
			if err != nil {
				return 0, err
			}
			if !more {
				end = true
			}
		}

		if end || stored == 5 {
			keyslice := keyword[:stored]
			keystring := string(keyslice)
			if !(keystring == "true" ||
				keystring == "false" ||
				keystring == "null") {
				return 0, ErrBadValue{string(keyword[:])}
			}

			return w.Write(keyslice)
		}
	}

	return written, nil
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

type ErrBadValue struct {
	Value string
}

func (e ErrBadValue) Error() string {
	return fmt.Sprintf("bad value: %s", e.Value)
}

func isSpace(c byte) bool {
	return c <= ' ' && (c == ' ' || c == '\t' || c == '\r' || c == '\n')
}

func SaneValueStart(c byte) bool {
	closer := closerFor(c)
	if closer != 0 {
		return true
	}
	return (c == 'n' || // for null
		c == 't' || // for true
		(c >= '1' && c <= '9') ||
		c == '-' ||
		c == 'f') // for false

}
