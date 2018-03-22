package types

import (
	"errors"

	"github.com/golang/protobuf/ptypes/empty"
)

var ErrResponseLengthMismatch = errors.New("Response length mismatch")
var ErrResponseStartTimeMismatch = errors.New("Response start time mismatch")
var ErrNotImplementedYet = errors.New("this feature is not implemented yet")
var ErrTimeoutExceeded = errors.New("timeout while fetching Response")
var ErrNonFatalErrors = errors.New("Response contains non-fatal errors")
var ErrNotFound = errors.New("metric not found")
var ErrNoResponseFetched = errors.New("No responses fetched from upstream")
var ErrNoMetricsFetched = errors.New("No metrics in the Response")

var ErrFailedToFetchFmt = "failed to fetch data from server group %v, code %v"

var EmptyMsg = &empty.Empty{}
