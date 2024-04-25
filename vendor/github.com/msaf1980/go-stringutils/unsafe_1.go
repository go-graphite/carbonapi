package stringutils

// compability with Golang 1.20 proposal https://github.com/golang/go/issues/53003
var (
	String     func(ptr *byte, length int) string = UnsafeStringFromPtr
	StringData func(str string) *byte             = UnsafeStringBytePtr
)
