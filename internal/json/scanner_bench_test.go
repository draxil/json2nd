package json

import (
	"testing"
)

var input = []byte(
	`
  "alpha" : {
    "zeeta" : [1,2,3,4,5,6,7,8,9,10],
    "cheese" : "cake",
  },
  "cheesy" : 1,
  "cheese": 4
}
`,
)

func BenchmarkScanForKey(b *testing.B) {
	for n := 0; n < b.N; n++ {
		s := NewScanState('{')
		s.seekFor("cheese")
		_, err := s.scan(input, 0, len(input))
		if err != nil {
			b.Fatal(err)
		}
	}
}
