package errors

import (
	"github.com/ansel1/merry"
)

var ErrBackendError = merry.New("error fetching data from backend")
var ErrResponseTypeMismatch = merry.New("type for the response doesn't match what's expected")
var ErrTimeout = merry.New("timeout while fetching data")
