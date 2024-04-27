package types

import (
	"errors"
	"net/http"

	"github.com/ansel1/merry/v2"

	"github.com/golang/protobuf/ptypes/empty"
)

var EmptyMsg = &empty.Empty{}

var ErrResponseTypeMismatch = errors.New("type for the response doesn't match what's expected")
var ErrResponseLengthMismatch = errors.New("response length mismatch")
var ErrResponseStartTimeMismatch = errors.New("response start time mismatch")
var ErrResponseStepTimeMismatch = errors.New("response step time mismatch")
var ErrNotImplementedYet = errors.New("this feature is not implemented yet")
var ErrNotSupportedByBackend = errors.New("this feature is not supported by backend")
var ErrForbidden = merry.Wrap(errors.New("forbidden"), merry.WithHTTPCode(http.StatusForbidden))
var ErrTimeoutExceeded = merry.Wrap(errors.New("timeout while fetching Response"),
	merry.WithHTTPCode(http.StatusGatewayTimeout))
var ErrNonFatalErrors = errors.New("response contains non-fatal errors")
var ErrNotFound = errors.New("metric not found")
var ErrNoResponseFetched = errors.New("no responses fetched from upstream")
var ErrNoMetricsFetched = merry.Wrap(errors.New("no metrics in the Response"), merry.WithHTTPCode(http.StatusNotFound))
var ErrMaxTriesExceeded = errors.New("max tries exceeded")
var ErrFailed = errors.New("failed due to error")
var ErrFailedToFetch = errors.New("failed to fetch data from server/group")
var ErrNoRequests = errors.New("no requests to fetch")
var ErrNoTagSpecified = errors.New("no tag specified")
var ErrNoServersSpecified = errors.New("no servers specified")
var ErrConcurrencyLimitNotSet = errors.New("concurrency limit is not set")
var ErrUnmarshalFailed = errors.New("unmarshal failed")
var ErrBackendError = merry.Wrap(errors.New("error fetching data from backend"),
	merry.WithHTTPCode(http.StatusServiceUnavailable))
var ErrResponceError = errors.New("error while fetching Response")

func ReturnNonNotFoundError(errs []error) []error {
	var errList []error
	for _, err := range errs {
		if !errors.Is(err, ErrNotFound) {
			errList = append(errList, err)
		}
	}
	return errList
}
