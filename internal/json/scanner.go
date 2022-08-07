package json

import (
	"fmt"
)

type state struct {
	in            byte
	closer        byte
	closerBalance int
	last          byte
	lastNotWs     byte
	inStr         bool
	key           bool
	open          bool
	seek          struct {
		keyName  []byte
		cursor   int
		matching bool
	}
	seeking   bool
	seekFound bool
}

func NewScanState(in byte) *state {

	var inStr bool
	if in == '"' {
		inStr = true
	}

	return &state{
		in:     in,
		inStr:  inStr,
		closer: closerFor(in),
		open:   true,
	}
}

func (s *state) seekFor(key string) {
	s.seek.keyName = []byte(key)
	s.seek.cursor = 0
	s.seek.matching = false
	s.seeking = true
}

func (s *state) scan(chunk []byte, idx, max int) (int, error) {

	if s.closer == 0 {
		return 0, ErrBadJSONValue{s.in}
	}

	for ; idx < max; idx++ {
		b := chunk[idx]

		// is whitespace?
		if b <= ' ' {
			if b == ' ' || b == '\t' || b == '\r' || (s.seeking && b == '\n') {
				s.last = b
				continue
			}

			// FUTURE: better way to communicate skips
			if !s.seeking && b == '\n' && !s.inStr {
				return idx, nil
			}
		}

		// start or end of a string:
		if b == '"' {
			if s.inStr && s.last != '\\' {
				// end of string.
				s.inStr = false
				if s.in == '{' && s.key {
					s.key = false
					if s.seeking {
						if s.closerBalance == 0 {
							if s.seek.matching && s.seek.cursor == len(s.seek.keyName) {
								s.seeking = false
								s.seekFound = true
								return idx, nil
							}
						}
						s.last = b
					}
				}
				if s.closer == '"' {
					s.last = b
					s.open = false
					return idx, nil
				}
			} else if !s.inStr {
				// start of string
				s.inStr = true
				if s.in == '{' && !s.key && s.lastNotWs != ':' {
					s.key = true
				}
				if s.seeking && s.closerBalance == 0 && s.key {
					s.seek.cursor = 0
					s.seek.matching = true
				}
			}
		} else if s.seeking && s.key && s.seek.matching {
			// is this the object key we were looking for?
			if s.seek.cursor >= len(s.seek.keyName) {
				s.seek.matching = false
			} else if s.seek.keyName[s.seek.cursor] != b {
				s.seek.matching = false
			} else {
				s.seek.cursor++
			}
		}

		// possibly the end of whatever we are scanning
		if !s.inStr {
			if b == s.closer {
				if s.closerBalance == 0 {
					s.open = false
					s.last = b
					s.lastNotWs = b
					return idx, nil
				} else {
					s.closerBalance--
				}

			} else if b == s.in {
				s.closerBalance++
			}
		}

		s.last = b
		s.lastNotWs = b
	}

	return max, nil
}

// TODO: we also have ErrBadJSONValue :)
type ErrBadJSONValue struct {
	Char byte
}

func (e ErrBadJSONValue) Error() string {
	return fmt.Sprintf("don't know how to process JSON value starting with: %c", e.Char)
}
