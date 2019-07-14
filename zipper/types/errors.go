package types

import (
	"github.com/ansel1/merry"

	"github.com/golang/protobuf/ptypes/empty"
)

var EmptyMsg = &empty.Empty{}

var ErrResponseLengthMismatch = merry.New("response length mismatch")
var ErrResponseStartTimeMismatch = merry.New("response start time mismatch")
var ErrResponseStepTimeMismatch = merry.New("response step time mismatch")
var ErrNotImplementedYet = merry.New("this feature is not implemented yet")
var ErrNotSupportedByBackend = merry.New("this feature is not supported by backend")
var ErrTimeoutExceeded = merry.New("timeout while fetching Response")
var ErrNonFatalErrors = merry.New("response contains non-fatal errors")
var ErrNotFound = merry.New("metric not found")
var ErrNoResponseFetched = merry.New("no responses fetched from upstream")
var ErrNoMetricsFetched = merry.New("no metrics in the Response")
var ErrMaxTriesExceeded = merry.New("max tries exceeded")
var ErrFailedToFetch = merry.New("failed to fetch data from server/group")
var ErrNoRequests = merry.New("no requests to fetch")
var ErrNoTagSpecified = merry.New("no tag specified")
var ErrNoServersSpecified = merry.New("no servers specified")
var ErrConcurrencyLimitNotSet = merry.New("concurrency limit is not set")

func ReturnNonNotFoundError(errors []merry.Error) []merry.Error {
	var errList []merry.Error
	for _, err := range errors {
		if !merry.Is(err, ErrNotFound) {
			errList = append(errList, err)
		}
	}
	return errList
}
