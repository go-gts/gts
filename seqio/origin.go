package seqio

import (
	"fmt"

	"github.com/go-gts/gts"
)

func toOriginLength(length int) int {
	lines := length / 60
	ret := lines * 76

	lastLine := length % 60

	if lastLine == 0 {
		return ret
	}

	blocks := lastLine / 10
	ret += 10 + blocks*11

	lastBlock := lastLine % 10
	if lastBlock == 0 {
		return ret
	}

	return ret + lastBlock + 1
}

func fromOriginLength(length int) int {
	lines := length / 76
	ret := lines * 60

	lastLine := length % 76
	if lastLine == 0 {
		return ret
	}

	lastLine -= 11
	blocks := lastLine / 11
	return ret + (blocks * 10) + (lastLine % 11)
}

// Origin represents a GenBank sequence origin value.
type Origin struct {
	Buffer []byte
	Parsed bool
}

// NewOrigin formats a byte slice into GenBank sequence origin format.
func NewOrigin(p []byte) *Origin {
	length := len(p)
	q := make([]byte, toOriginLength(length))
	offset := 0
	for i := 0; i < length; i += 60 {
		prefix := fmt.Sprintf("%9d", i+1)
		offset += copy(q[offset:], prefix)
		for j := 0; j < 60 && i+j < length; j += 10 {
			start := i + j
			end := gts.Min(i+j+10, length)
			q[offset] = spaceByte
			offset++
			offset += copy(q[offset:], p[start:end])
		}
		q[offset] = '\n'
		offset++
	}
	return &Origin{q, false}
}

// Bytes converts the GenBank sequence origin into a byte slice.
func (o *Origin) Bytes() []byte {
	if !o.Parsed {
		p := o.Buffer
		if len(p) < 12 {
			return nil
		}

		length := fromOriginLength(len(p))
		q := make([]byte, length)
		offset, start := 0, 0
		for i := 0; i < length; i += 60 {
			start += 9
			for j := 0; j < 60 && i+j < length; j += 10 {
				start++
				end := gts.Min(start+10, len(p)-1)
				offset += copy(q[offset:], p[start:end])
				start = end
			}
			start++
		}

		o.Buffer = q
		o.Parsed = true
	}

	return o.Buffer
}

// String satisfies the fmt.Stringer interface.
func (o Origin) String() string {
	if !o.Parsed {
		return string(o.Buffer)
	}
	return string(NewOrigin(o.Buffer).Buffer)
}

// Len returns the actual sequence length.
func (o Origin) Len() int {
	if len(o.Buffer) == 0 {
		return 0
	}
	if o.Parsed {
		return len(o.Buffer)
	}
	return fromOriginLength(len(o.Buffer))
}
