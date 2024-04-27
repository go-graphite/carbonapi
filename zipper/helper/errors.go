package helper

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/ansel1/merry/v2"

	"github.com/go-graphite/carbonapi/pkg/parser"
	"github.com/go-graphite/carbonapi/zipper/types"
)

func RequestError(err error, server string) error {
	// with code InternalServerError by default, overwritten by custom error
	if errors.Is(err, context.DeadlineExceeded) {
		return merry.Wrap(types.ErrTimeoutExceeded,
			merry.WithValue("server", server),
			merry.WithCause(err),
		)
	}
	if urlErr, ok := err.(*url.Error); ok {
		if netErr, ok := urlErr.Err.(*net.OpError); ok {
			return merry.Wrap(types.ErrBackendError,
				merry.WithValue("server", server),
				merry.WithCause(netErr),
			)
		}
	}
	if netErr, ok := err.(*net.OpError); ok {
		return merry.Wrap(types.ErrBackendError,
			merry.WithValue("server", server),
			merry.WithCause(netErr),
		)
	}
	return merry.Wrap(types.ErrResponceError,
		merry.WithValue("server", server),
	)
}

func HttpErrorCode(err error) (code int) {
	if err == nil {
		code = http.StatusOK
	} else {
		c := merry.Wrap(err)
		if c == nil {
			c = err
		}

		code = merry.HTTPCode(err)
		if code == http.StatusNotFound {
			return
		} else if code == http.StatusInternalServerError && errors.Is(c, parser.ErrInvalidArg) {
			// check for invalid args, see applyByNode rewrite function
			code = http.StatusBadRequest
		}

		if code == http.StatusGatewayTimeout || code == http.StatusBadGateway || errors.Is(c, types.ErrFailedToFetch) {
			// simplify code, one error type for communications errors, all we can retry
			code = http.StatusServiceUnavailable
		}
	}

	return
}

// for stable return code on multiply errors
func recalcCode(code, newCode int) int {
	if newCode == http.StatusGatewayTimeout || newCode == http.StatusBadGateway {
		// simplify code, one error type for communications errors, all we can retry
		newCode = http.StatusServiceUnavailable
	}
	if code == 0 || code == http.StatusNotFound {
		return newCode
	}

	if newCode >= 400 && newCode < 500 && code >= 400 && code < 500 {
		if newCode == http.StatusBadRequest {
			return newCode
		} else if newCode == http.StatusForbidden && code != http.StatusBadRequest {
			return newCode
		}
	}
	if newCode < code {
		code = newCode
	}
	return code
}

func RootCause(err error) error {
	for {
		cause := merry.Cause(err)
		if cause == nil {
			return err
		}
		err = cause
	}
}

// MerryRootError strip merry error chain
func MerryRootError(err error) string {
	c := RootCause(err)
	if c == nil {
		c = err
	}
	return merryError(c)
}

func merryError(err error) string {
	if msg := err.Error(); len(msg) > 0 {
		return strings.TrimRight(msg, "\n")
	} else {
		return err.Error()
	}
}

func MergeHttpErrors(errs []error) (int, []string) {
	returnCode := http.StatusNotFound
	errMsgs := make([]string, 0)
	for _, err := range errs {
		c := RootCause(err)
		if c == nil {
			c = err
		}

		code := merry.HTTPCode(err)
		if code == http.StatusNotFound {
			continue
		} else if code == http.StatusInternalServerError && errors.Is(c, parser.ErrInvalidArg) {
			// check for invalid args, see applyByNode rewrite function
			code = http.StatusBadRequest
		}

		errMsgs = append(errMsgs, merryError(c))

		returnCode = recalcCode(returnCode, code)
	}

	return returnCode, errMsgs
}

func MergeHttpErrorMap(errorsMap map[string]error) (returnCode int, errMap map[string]string) {
	returnCode = http.StatusNotFound
	errMap = make(map[string]string)
	for key, err := range errorsMap {
		c := RootCause(err)
		if c == nil {
			c = err
		}

		code := merry.HTTPCode(err)
		if code == http.StatusNotFound {
			continue
		} else if code == http.StatusInternalServerError && errors.Is(c, parser.ErrInvalidArg) {
			// check for invalid args, see applyByNode rewrite function
			code = http.StatusBadRequest
		}

		msg := merryError(c)
		errMap[key] = msg
		returnCode = recalcCode(returnCode, code)
	}

	return
}

func HttpErrorByCode(err error) error {
	var returnErr error
	if err == nil {
		returnErr = types.ErrNoMetricsFetched
	} else {
		code := merry.HTTPCode(err)
		msg := stripHtmlTags(err.Error(), 0)
		if code == http.StatusForbidden {
			returnErr = types.ErrForbidden
			if len(msg) > 0 {
				// pass message to caller
				returnErr = merry.Wrap(returnErr, merry.WithMessage(msg))
			}
		} else if code == http.StatusServiceUnavailable || code == http.StatusBadGateway || code == http.StatusGatewayTimeout {
			returnErr = merry.Wrap(types.ErrFailedToFetch,
				merry.WithHTTPCode(code),
				merry.WithMessage(msg),
			)
		} else {
			returnErr = merry.Wrap(types.ErrFailed,
				merry.WithHTTPCode(code),
				merry.WithMessage(msg),
			)
		}
	}

	return returnErr
}
