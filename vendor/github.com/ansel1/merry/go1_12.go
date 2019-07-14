// +build !go1.13

package merry

import (
	errors "golang.org/x/xerrors"
)

// If using <go1.13, polyfill errors.Is/As with golang.org/x/xerrors
// xerrors can be removed once <go1.12 support is dropped

// implements Is by delegating to errors
var is = errors.Is

// implements As by delegating to errors
var as = errors.As
