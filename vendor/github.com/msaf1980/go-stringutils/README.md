# stringutils

Some string utils

`UnsafeString([]byte) string` return unsafe string from bytes slice indirectly (without allocation)
`UnsafeStringFromPtr(*byte, length) string` return unsafe string from bytes slice pointer indirectly (without allocation)
`UnsafeStringBytes(*string) []bytes` return unsafe string bytes indirectly (without allocation)

`Split2(s string, sep string) (string, string, int)`  Split2 return the split string results (without memory allocations). Use Index for find separator.
`SplitN(s string, sep string, buf []string) ([]string, int)` // SplitN return splitted slice (use pre-allocated buffer) and end position (for detect if string contains more fields for split). Use Index for find separator.


`Builder` very simular to strings.Builder, but has better perfomance (at golang 1.14).

`Template` is a simple templating system
