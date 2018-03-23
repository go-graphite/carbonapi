package types

import (
	"errors"

	"github.com/golang/protobuf/ptypes/empty"
)

var ErrResponseLengthMismatch = errors.New("response length mismatch")
var ErrResponseStartTimeMismatch = errors.New("response start time mismatch")
var ErrNotImplementedYet = errors.New("this feature is not implemented yet")
var ErrTimeoutExceeded = errors.New("timeout while fetching Response")
var ErrNonFatalErrors = errors.New("response contains non-fatal errors")
var ErrNotFound = errors.New("metric not found")
var ErrNoResponseFetched = errors.New("no responses fetched from upstream")
var ErrNoMetricsFetched = errors.New("no metrics in the Response")
var ErrMaxTriesExceeded = errors.New("max tries exceeded")

var ErrFailedToFetchFmt = "failed to fetch data from server group %v, code %v"

var EmptyMsg = &empty.Empty{}
