package stringutils

import "io"

// WriteString writes the contents of the string s to w, which accepts a slice of bytes. No bytes alloation instead of io.WriteString.
func WriteString(w io.Writer, s string) (n int, err error) {
	return w.Write(UnsafeStringBytes(&s))
}
