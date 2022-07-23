package json

// TODO: MASTER OFFSET FOR ERRORS?

import (
	"fmt"
	"io"
	"strings"
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

	closer := closerFor(start)
	closerBalance := 0
	if closer == 0 {
		return false, ErrInternal{fmt.Errorf("no closer for: %c", start)}
	}
	onkey := false
	inStr := false
	last := byte(0)
	keybuf := strings.Builder{}
	// TODO: SHARED
	// TODO: NOT BIG AN GUGLY
	for {
		for ; j.idx < len(j.buf); j.idx++ {
			b := j.buf[j.idx]

			if start != closer {
				if inStr && b != '"' {
					if onkey {
						keybuf.WriteByte(b)
					}
					last = b
					continue
				}
				if b == '"' {
					if !inStr {
						inStr = true
						last = b

						if onkey {
							onkey = false
						}
						if !onkey {
							onkey = true
						}

						continue
					}
					if inStr && last != '\\' {
						inStr = false

						if onkey {
							key := keybuf.String()
							if key == k {
								return true, nil
							}
							keybuf.Reset()
						}

						last = b
						continue
					}
				}
			}

			if b == closer {
				if closerBalance > 0 {
					closerBalance--
					continue
				}
				return false, nil
			}
			if b == start {
				closerBalance++
			}
		}

		more, err := j.data()
		if err != nil {
			// TODO: context?
			return false, err
		}
		if !more {
			break
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
	startIdx := j.idx
	start := j.buf[j.idx]
	j.idx++

	if start == 0 {
		return 0, errNoObject()
	}

	if !includeDeliminators {
		startIdx++
	}

	closer := closerFor(start)
	if closer == 0 {
		return 0, ErrInternal{fmt.Errorf("no closer for: %c", start)}
	}

	end := -1
	closerBalance := 0
	alreadyWritten := 0
	inStr := false
	last := byte(0)

	// TODO: LESS BIG AND DUMB
	for {
		for ; j.idx < len(j.buf); j.idx++ {
			b := j.buf[j.idx]

			// TODO: FOR THE SANE VERSION NOT SO NESTED
			if start != closer {
				if inStr && b != '"' {
					last = b
					continue
				}
				if b == '"' {
					if !inStr {
						inStr = true
						last = b
						continue
					}
					if inStr && last != '\\' {
						inStr = false
						last = b
						continue
					}
				}

				if b == closer {
					if closerBalance > 0 {
						closerBalance--
						continue
					}
					last = b
					end = j.idx
					break
				}
				if b == start {
					closerBalance++
				}
			}

			if start == closer {
				if last == '\\' {
					last = b
					continue
				}
				if b == closer {
					last = b
					end = j.idx
					break
				}
			}

			last = b
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

	if !includeDeliminators {
		end--
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
