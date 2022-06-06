# stringutils

Some string utils

`UnsafeString([]byte) string` return unsafe string from bytes slice indirectly (without allocation)

`UnsafeStringFromPtr(*byte, length) string` return unsafe string from bytes slice pointer indirectly (without allocation)

`UnsafeStringBytes(*string) []bytes` return unsafe string bytes indirectly (without allocation)


`Split2(s string, sep string) (string, string, int)`  Split2 return the split string results (without memory allocations). Use Index for find separator.

`SplitN(s string, sep string, buf []string) ([]string, int)` // SplitN return splitted slice (use pre-allocated buffer) and end position (for detect if string contains more fields for split). Use Index for find separator.

`Reverse(string) string` return reversed string (rune-wise left to right)
`ReverseSegments(string, delim) string` return reversed string by segments around string delimiter (`ReverseSegments("hello, world", ", ")` return `world, hello`).

`Replace(s, old, new string, n int) (string, changed)` // Replace returns a copy of the string s with the first n non-overlapping instances of old replaced by new. Also return change flag.

`ReplaceAll(s, old, new string) (string, changed)` // Replace returns a copy of the string s with all non-overlapping instances of old replaced  by new. Also return change flag.

`WriteString(w io.Writer, s string) (int, error)` writes the contents of the string s to w, which accepts a slice of bytes. No bytes alloation instead of io.WriteString.

`Builder` very simular to strings.Builder, but has better perfomance in some cases (reallocate with scale 2, if needed, also append numbers in-place) (at golang 1.14).

`Template` is a simple templating system
