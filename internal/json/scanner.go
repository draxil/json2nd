package json

import (
	"bytes"
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
	seek          []byte
	seeking       bool
	seekFound     bool
	keybuf        bytes.Buffer
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
	s.seek = []byte(key)
	s.seeking = true
}

func (s *state) scan(chunk []byte, idx, max int) (int, error) {

	if s.closer == 0 {
		return 0, ErrBadJSONValue{s.in}
	}

	for ; idx < max; idx++ {
		b := chunk[idx]

		// is whitespace?
		if b <= ' ' && (b == ' ' || b == '\t' || b == '\r' || b == '\n') {
			s.last = b
			continue
		}

		if b == '"' {
			if s.inStr && s.last != '\\' {
				s.inStr = false
				if s.in == '{' && s.key {
					s.key = false
					if s.seeking {
						if s.closerBalance == 0 {
							// TODO: probably not this way
							if string(s.keybuf.Bytes()) == string(s.seek) {
								s.seeking = false
								s.seekFound = true
								return idx, nil
							}
						}
						s.keybuf.Reset()
						s.last = b
					}
				}
				if s.closer == '"' {
					s.last = b
					s.open = false
					return idx, nil
				}
			} else if !s.inStr {
				s.inStr = true
				if s.in == '{' && !s.key && s.lastNotWs != ':' {
					s.key = true
				}
			}
		} else if s.seeking && s.key {
			// TODO: maybe don't care if not matching so far or something?
			s.keybuf.WriteByte(b)
		}

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

type ErrBadJSONValue struct {
	Char byte
}

func (e ErrBadJSONValue) Error() string {
	return fmt.Sprintf("don't know how to process JSON value starting with: %c", e.Char)
}
