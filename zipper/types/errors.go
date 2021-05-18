package types

import (
	"net/http"

	"github.com/ansel1/merry"

	"github.com/golang/protobuf/ptypes/empty"
)

var EmptyMsg = &empty.Empty{}

var ErrResponseTypeMismatch = merry.New("type for the response doesn't match what's expected")
var ErrResponseLengthMismatch = merry.New("response length mismatch")
var ErrResponseStartTimeMismatch = merry.New("response start time mismatch")
var ErrResponseStepTimeMismatch = merry.New("response step time mismatch")
var ErrNotImplementedYet = merry.New("this feature is not implemented yet")
var ErrNotSupportedByBackend = merry.New("this feature is not supported by backend")
var ErrForbidden = merry.New("forbidden").WithHTTPCode(http.StatusForbidden)
var ErrTimeoutExceeded = merry.New("timeout while fetching Response").WithHTTPCode(http.StatusGatewayTimeout)
var ErrNonFatalErrors = merry.New("response contains non-fatal errors")
var ErrNotFound = merry.New("metric not found")
var ErrNoResponseFetched = merry.New("no responses fetched from upstream")
var ErrNoMetricsFetched = merry.New("no metrics in the Response").WithHTTPCode(http.StatusNotFound)
var ErrMaxTriesExceeded = merry.New("max tries exceeded")
var ErrFailed = merry.New("failed due to error")
var ErrFailedToFetch = merry.New("failed to fetch data from server/group")
var ErrNoRequests = merry.New("no requests to fetch")
var ErrNoTagSpecified = merry.New("no tag specified")
var ErrNoServersSpecified = merry.New("no servers specified")
var ErrConcurrencyLimitNotSet = merry.New("concurrency limit is not set")
var ErrUnmarshalFailed = merry.New("unmarshal failed")
var ErrBackendError = merry.New("error fetching data from backend").WithHTTPCode(http.StatusServiceUnavailable)
var ErrResponceError = merry.New("error while fetching Response")

func ReturnNonNotFoundError(errors []merry.Error) []merry.Error {
	var errList []merry.Error
	for _, err := range errors {
		if !merry.Is(err, ErrNotFound) {
			errList = append(errList, err)
		}
	}
	return errList
}
