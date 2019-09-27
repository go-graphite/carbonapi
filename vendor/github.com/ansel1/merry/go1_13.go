// +build go1.13

package merry

import "errors"

// If using >=go1.13, golang.org/x/xerrors is not needed.
// xerrors can be removed once <go1.12 support is dropped

// implements Is by delegating to errors
var is = errors.Is

// implements As by delegating to errors
var as = errors.As
