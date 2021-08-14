package stringutils

import (
	"reflect"
	"unsafe"
)

// UnsafeString returns the string under byte buffer
func UnsafeString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// UnsafeStringFromPtr returns the string with specific length under byte buffer
func UnsafeStringFromPtr(ptr *byte, length int) (s string) {
	str := (*reflect.StringHeader)(unsafe.Pointer(&s))
	str.Data = uintptr(unsafe.Pointer(ptr))
	str.Len = length

	return s
}

// UnsafeStringBytes returns the string bytes
func UnsafeStringBytes(s *string) []byte {
	return *(*[]byte)(unsafe.Pointer((*reflect.SliceHeader)(unsafe.Pointer(s))))
}
