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
	return
}

// UnsafeStringBytes returns the string bytes
func UnsafeStringBytes(s *string) []byte {
	return *(*[]byte)(unsafe.Pointer((*reflect.SliceHeader)(unsafe.Pointer(s))))
}

// UnsafeStringBytePtr returns the string byte ptr
func UnsafeStringBytePtr(s string) *byte {
	return &(*(*[]byte)(unsafe.Pointer((*reflect.SliceHeader)(unsafe.Pointer(&s)))))[0]
}

func UnsafeBytes(ptr *byte, length, cap int) (b []byte) {
	bs := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	bs.Data = uintptr(unsafe.Pointer(ptr))
	bs.Len = length
	bs.Cap = cap
	return *(*[]byte)(unsafe.Pointer(&b))
}
