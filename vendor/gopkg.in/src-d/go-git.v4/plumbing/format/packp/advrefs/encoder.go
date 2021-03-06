package advrefs

import (
	"bytes"
	"io"
	"sort"

	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/format/packp"
	"gopkg.in/src-d/go-git.v4/plumbing/format/packp/pktline"
)

// An Encoder writes AdvRefs values to an output stream.
type Encoder struct {
	data *AdvRefs         // data to encode
	pe   *pktline.Encoder // where to write the encoded data
	err  error            // sticky error
}

// NewEncoder returns a new encoder that writes to w.
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{
		pe: pktline.NewEncoder(w),
	}
}

// Encode writes the AdvRefs encoding of v to the stream.
//
// All the payloads will end with a newline character.  Capabilities,
// references and shallows are writen in alphabetical order, except for
// peeled references that always follow their corresponding references.
func (e *Encoder) Encode(v *AdvRefs) error {
	e.data = v

	for state := encodePrefix; state != nil; {
		state = state(e)
	}

	return e.err
}

type encoderStateFn func(*Encoder) encoderStateFn

func encodePrefix(e *Encoder) encoderStateFn {
	for _, p := range e.data.Prefix {
		if bytes.Equal(p, pktline.Flush) {
			if e.err = e.pe.Flush(); e.err != nil {
				return nil
			}
			continue
		}
		if e.err = e.pe.Encodef("%s\n", string(p)); e.err != nil {
			return nil
		}
	}

	return encodeFirstLine
}

// Adds the first pkt-line payload: head hash, head ref and capabilities.
// Also handle the special case when no HEAD ref is found.
func encodeFirstLine(e *Encoder) encoderStateFn {
	head := formatHead(e.data.Head)
	separator := formatSeparator(e.data.Head)
	capabilities := formatCaps(e.data.Capabilities)

	if e.err = e.pe.Encodef("%s %s\x00%s\n", head, separator, capabilities); e.err != nil {
		return nil
	}

	return encodeRefs
}

func formatHead(h *plumbing.Hash) string {
	if h == nil {
		return plumbing.ZeroHash.String()
	}

	return h.String()
}

func formatSeparator(h *plumbing.Hash) string {
	if h == nil {
		return noHead
	}

	return head
}

func formatCaps(c *packp.Capabilities) string {
	if c == nil {
		return ""
	}

	c.Sort()

	return c.String()
}

// Adds the (sorted) refs: hash SP refname EOL
// and their peeled refs if any.
func encodeRefs(e *Encoder) encoderStateFn {
	refs := sortRefs(e.data.References)
	for _, r := range refs {
		hash, _ := e.data.References[r]
		if e.err = e.pe.Encodef("%s %s\n", hash.String(), r); e.err != nil {
			return nil
		}

		if hash, ok := e.data.Peeled[r]; ok {
			if e.err = e.pe.Encodef("%s %s^{}\n", hash.String(), r); e.err != nil {
				return nil
			}
		}
	}

	return encodeShallow
}

func sortRefs(m map[string]plumbing.Hash) []string {
	ret := make([]string, 0, len(m))
	for k := range m {
		ret = append(ret, k)
	}
	sort.Strings(ret)

	return ret
}

// Adds the (sorted) shallows: "shallow" SP hash EOL
func encodeShallow(e *Encoder) encoderStateFn {
	sorted := sortShallows(e.data.Shallows)
	for _, hash := range sorted {
		if e.err = e.pe.Encodef("shallow %s\n", hash); e.err != nil {
			return nil
		}
	}

	return encodeFlush
}

func sortShallows(c []plumbing.Hash) []string {
	ret := []string{}
	for _, h := range c {
		ret = append(ret, h.String())
	}
	sort.Strings(ret)

	return ret
}

func encodeFlush(e *Encoder) encoderStateFn {
	e.err = e.pe.Flush()
	return nil
}
