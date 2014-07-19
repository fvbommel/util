// Package rope implements a "heavy-weight string", which represents very long
// strings more efficiently (especially when many concatenations are performed).
//
// It may also need less memory if it contains repeated substrings, or if you
// use several large strings that are similar to each other.
//
// Rope values are immutable, so each operation returns its result instead
// of modifying the receiver. This immutability also makes them thread-safe.
package rope

import (
	"bytes"
	"fmt"
	"io"
)

var emptyRope = Rope{emptyNode} // A Rope containing the empty node.

// Rope implements a non-contiguous string.
// The zero value is an empty rope.
type Rope struct {
	node node // The root node of this rope. May be nil.
}

// New returns a Rope representing a given string.
func New(arg string) Rope {
	if len(arg) == 0 {
		return emptyRope
	}
	return Rope{
		node: leaf(arg),
	}
}

// String materializes the Rope as a string value.
func (r Rope) String() string {
	if r.node == nil {
		return ""
	}
	// In the trivial case, avoid allocation
	if l, ok := r.node.(leaf); ok {
		return string(l)
	}
	// The rope is not contiguous.
	return string(r.Bytes())
}

// GoString materializes the Rope as a quoted string value.
func (r Rope) GoString() string {
	// Perhaps technically more correct, but not nearly as useful:
	//	 return fmt.Sprintf("rope.New(%q)", r.String())

	// Instead, (mostly) pretend we're a regular string.
	if MarkGoStringedRope {
		return fmt.Sprintf("/*Rope*/ %#v", r.String())
	}
	return fmt.Sprintf("%#v", r.String())
}

// Bytes returns the string represented by this Rope as a []byte.
func (r Rope) Bytes() []byte {
	len := r.Len()
	if len == 0 {
		return nil
	}
	buf := bytes.NewBuffer(make([]byte, 0, len))
	r.WriteTo(buf)
	return buf.Bytes()
}

// WriteTo writes the value of this Rope to the provided writer.
func (r Rope) WriteTo(w io.Writer) (n int64, err error) {
	if r.node == nil {
		return 0, nil // Nothing to do
	}
	return r.node.WriteTo(w)
}

// Len returns the length of the string represented by the Rope.
func (r Rope) Len() int64 {
	if r.node == nil {
		return 0
	}
	return r.node.length()
}

// Append returns the Rope representing the arguments appended to this rope.
func (r Rope) Append(rhs ...Rope) Rope {
	// Handle nil-node receiver
	for r.node == nil && len(rhs) > 0 {
		r = rhs[0]
		rhs = rhs[1:]
	}
	if len(rhs) == 0 {
		return r
	}

	list := make([]node, 0, len(rhs))
	for _, item := range rhs {
		if item.node != nil {
			list = append(list, item.node)
		}
	}
	node := concMany(r.node, list...)
	return Rope{node: node}
}

// DropPrefix returns a postfix of a rope, starting at index.
// It's analogous to str[start:].
//
// If start >= r.Len(), an empty Rope is returned.
func (r Rope) DropPrefix(start int64) Rope {
	if start <= 0 || r.node == nil {
		return r
	}
	return Rope{
		node: r.node.dropPrefix(start),
	}
}

// DropPostfix returns the prefix of a rope ending at end.
// It's analogous to str[:end].
//
// If end <= 0, an empty Rope is returned.
func (r Rope) DropPostfix(end int64) Rope {
	if r.node == nil {
		return r
	}
	return Rope{
		node: r.node.dropPostfix(end),
	}
}

// Slice returns the substring of a Rope, analogous to str[start:end].
// It is equivalent to r.DropPostfix(end).DropPrefix(start).
//
// If start >= end, start >= r.Len() or end == 0, an empty Rope is returned.
func (r Rope) Slice(start, end int64) Rope {
	if r.node == nil {
		return r
	}
	if start < 0 {
		start = 0
	}
	if start >= end {
		return emptyRope
	}
	return Rope{node: r.node.slice(start, end)}
}
