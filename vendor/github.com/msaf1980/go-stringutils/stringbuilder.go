package stringutils

import (
	"unicode/utf8"
)

// A Builder is used to efficiently build a string using Write methods (with better perfomance than strings.Builder).
// It minimizes memory copying. The zero value is ready to use.
// Do not copy a non-zero Builder.
type Builder struct {
	data   []byte
	length int
}

// grow scale factor for needed resize
const scaleFactor = 2

// Len returns the number of accumulated bytes; b.Len() == len(b.String()).
func (sb *Builder) Len() int {
	return sb.length
}

// Cap returns the capacity of the builder's underlying byte slice. It is the
// total space allocated for the string being built and includes any bytes
// already written.
func (sb *Builder) Cap() int {
	return cap(sb.data)
}

// Bytes returns the accumulated bytes.
func (sb *Builder) Bytes() []byte {
	return sb.data[0:sb.length]
}

// String returns the accumulated string.
func (sb *Builder) String() string {
	if sb.length == 0 {
		return ""
	}
	return UnsafeStringFromPtr(&sb.data[0], sb.length)
}

// Grow grows b's capacity, if necessary, to guarantee space for
// another n bytes. After Grow(n), at least n bytes can be written to b
// without another allocation.
func (sb *Builder) Grow(capacity int) {
	if capacity > sb.Cap() {
		b := make([]byte, capacity)
		copy(b, sb.data[:sb.length])
		sb.data = b
	}
}

// Reset resets the Builder to be empty.
func (sb *Builder) Reset() {
	if sb.length > 0 {
		sb.length = 0
	}
}

// Release resets the Builder to be empty and free buffer
func (sb *Builder) Release() {
	if cap(sb.data) > 0 {
		sb.data = nil
	}
	sb.Reset()
}

// Write like WriteBytes, but realized io.Writer interface
func (sb *Builder) Write(bytes []byte) (int, error) {
	if len(bytes) == 0 {
		return 0, nil
	}
	newlen := sb.length + len(bytes)
	if newlen > cap(sb.data) {
		scaled := sb.length * scaleFactor
		if newlen > scaled {
			sb.Grow(newlen)
		} else {
			sb.Grow(scaled)
		}
	}
	copy(sb.data[sb.length:], bytes)
	sb.length += len(bytes)

	return len(bytes), nil
}

// WriteBytes appends the contents of p to b's buffer.
func (sb *Builder) WriteBytes(bytes []byte) {
	if len(bytes) == 0 {
		return
	}
	newlen := sb.length + len(bytes)
	if newlen > cap(sb.data) {
		scaled := sb.length * scaleFactor
		if newlen > scaled {
			sb.Grow(newlen)
		} else {
			sb.Grow(scaled)
		}
	}
	copy(sb.data[sb.length:], bytes)
	sb.length += len(bytes)
}

// WriteString appends the contents of s to b's buffer.
func (sb *Builder) WriteString(s string) (int, error) {
	if len(s) == 0 {
		return 0, nil
	}
	newlen := sb.length + len(s)
	if newlen > cap(sb.data) {
		scaled := sb.length * scaleFactor
		if newlen > scaled {
			sb.Grow(newlen)
		} else {
			sb.Grow(scaled)
		}
	}
	copy(sb.data[sb.length:], s)
	sb.length += len(s)

	return len(s), nil
}

// WriteByte appends the byte c to b's buffer.
func (sb *Builder) WriteByte(c byte) error {
	if sb.length == cap(sb.data) {
		if sb.length == 0 {
			sb.Grow(2 * scaleFactor)
		} else {
			sb.Grow(sb.length * scaleFactor)
		}
	}
	sb.data[sb.length] = c
	sb.length++

	return nil
}

// WriteRune appends the UTF-8 encoding of Unicode code point r to b's buffer.
func (sb *Builder) WriteRune(r rune) (int, error) {
	if r < utf8.RuneSelf {
		sb.WriteByte(byte(r))

		return 1, nil
	} else {
		if sb.length+utf8.UTFMax > cap(sb.data) {
			if sb.length > 2*utf8.UTFMax {
				sb.Grow(sb.length * scaleFactor)
			} else {
				sb.Grow(sb.length + utf8.UTFMax*scaleFactor)
			}
		}
		length := utf8.EncodeRune(sb.data[sb.length:sb.length+utf8.UTFMax], r)
		sb.length += length

		return length, nil
	}
}

// Flush fake makethod for combatibility with buffered writer
func (sb *Builder) Flush() error {
	return nil
}
