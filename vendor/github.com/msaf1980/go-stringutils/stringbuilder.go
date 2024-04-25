package stringutils

import (
	"unicode/utf8"
)

// A Builder is used to efficiently build a string using Write methods (with better perfomance than strings.Builder).
// It minimizes memory copying. The zero value is ready to use.
// Do not copy a non-zero Builder.
type Builder struct {
	data []byte
}

// grow scale factor for needed resize
const scaleFactor = 2

// Len returns the number of accumulated bytes; b.Len() == len(b.String()).
func (sb *Builder) Len() int {
	return len(sb.data)
}

// Cap returns the capacity of the builder's underlying byte slice. It is the
// total space allocated for the string being built and includes any bytes
// already written.
func (sb *Builder) Cap() int {
	return cap(sb.data)
}

// Bytes returns the accumulated bytes.
func (sb *Builder) Bytes() []byte {
	return sb.data
}

// String returns the accumulated string.
func (sb *Builder) String() string {
	if len(sb.data) == 0 {
		return ""
	}
	return UnsafeStringFromPtr(&sb.data[0], len(sb.data))
}

// Grow grows b's capacity, if necessary, to guarantee space for
// another n bytes. After Grow(n), at least n bytes can be written to b
// without another allocation.
func (sb *Builder) Grow(capacity int) {
	if capacity > sb.Cap() {
		b := make([]byte, capacity)
		copy(b, sb.data)
		length := len(sb.data)
		sb.data = b[:length]
	}
}

// Reset resets the Builder to be empty.
func (sb *Builder) Reset() {
	if len(sb.data) > 0 {
		sb.data = sb.data[:0]
	}
}

// Truncate descrease the Builder length (dangerouse for partually truncated UTF strings).
func (sb *Builder) Truncate(length int) {
	if len(sb.data) > length {
		sb.data = sb.data[:length]
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
	sb.data = append(sb.data, bytes...)

	return len(bytes), nil
}

// WriteBytes appends the contents of p to b's buffer.
func (sb *Builder) WriteBytes(bytes []byte) {
	if len(bytes) == 0 {
		return
	}
	sb.data = append(sb.data, bytes...)
}

// WriteString appends the contents of s to b's buffer.
func (sb *Builder) WriteString(s string) (int, error) {
	if len(s) == 0 {
		return 0, nil
	}
	sb.data = append(sb.data, s...)

	return len(s), nil
}

// WriteByte appends the byte c to b's buffer.
func (sb *Builder) WriteByte(c byte) error {
	if len(sb.data) == 0 {
		sb.Grow(2 * scaleFactor)
	} else if len(sb.data) == cap(sb.data) {
		sb.Grow(len(sb.data) * scaleFactor)
	}
	length := len(sb.data)
	sb.data = sb.data[:length+1]
	sb.data[length] = c

	return nil
}

// WriteRune appends the UTF-8 encoding of Unicode code point r to b's buffer.
func (sb *Builder) WriteRune(r rune) (int, error) {
	if r < utf8.RuneSelf {
		sb.WriteByte(byte(r))

		return 1, nil
	} else {
		length := len(sb.data)
		n := length + utf8.UTFMax
		if n > cap(sb.data) {
			if length > 2*utf8.UTFMax {
				sb.Grow(length * scaleFactor)
			} else {
				sb.Grow(length + utf8.UTFMax*scaleFactor)
			}
		}
		sb.data = sb.data[:n]
		n = utf8.EncodeRune(sb.data[length:], r)
		sb.data = sb.data[:length+n]

		return length, nil
	}
}

// Flush fake makethod for combatibility with buffered writer
func (sb *Builder) Flush() error {
	return nil
}
