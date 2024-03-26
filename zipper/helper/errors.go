package helper

import (
	"context"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/ansel1/merry"
	"github.com/go-graphite/carbonapi/pkg/parser"
	"github.com/go-graphite/carbonapi/zipper/types"
)

func requestError(err error, server string) merry.Error {
	// with code InternalServerError by default, overwritten by custom error
	if merry.Is(err, context.DeadlineExceeded) {
		return types.ErrTimeoutExceeded.WithValue("server", server).WithCause(err)
	}
	if urlErr, ok := err.(*url.Error); ok {
		if netErr, ok := urlErr.Err.(*net.OpError); ok {
			return types.ErrBackendError.WithValue("server", server).WithCause(netErr)
		}
	}
	if netErr, ok := err.(*net.OpError); ok {
		return types.ErrBackendError.WithValue("server", server).WithCause(netErr)
	}
	return types.ErrResponceError.WithValue("server", server)
}

func HttpErrorCode(err merry.Error) (code int) {
	if err == nil {
		code = http.StatusOK
	} else {
		c := merry.RootCause(err)
		if c == nil {
			c = err
		}

		code = merry.HTTPCode(err)
		if code == http.StatusNotFound {
			return
		} else if code == http.StatusInternalServerError && merry.Is(c, parser.ErrInvalidArg) {
			// check for invalid args, see applyByNode rewrite function
			code = http.StatusBadRequest
		}

		if code == http.StatusGatewayTimeout || code == http.StatusBadGateway || merry.Is(c, types.ErrFailedToFetch) {
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

// MerryRootError strip merry error chain
func MerryRootError(err error) string {
	c := merry.RootCause(err)
	if c == nil {
		c = err
	}
	return merryError(c)
}

func merryError(err error) string {
	if msg := merry.Message(err); len(msg) > 0 {
		return strings.TrimRight(msg, "\n")
	} else {
		return err.Error()
	}
}

func MergeHttpErrors(errors []merry.Error) (int, []string) {
	returnCode := http.StatusNotFound
	errMsgs := make([]string, 0)
	for _, err := range errors {
		c := merry.RootCause(err)
		if c == nil {
			c = err
		}

		code := merry.HTTPCode(err)
		if code == http.StatusNotFound {
			continue
		} else if code == http.StatusInternalServerError && merry.Is(c, parser.ErrInvalidArg) {
			// check for invalid args, see applyByNode rewrite function
			code = http.StatusBadRequest
		}

		errMsgs = append(errMsgs, merryError(c))

		returnCode = recalcCode(returnCode, code)
	}

	return returnCode, errMsgs
}

func MergeHttpErrorMap(errorsMap map[string]merry.Error) (returnCode int, errMap map[string]string) {
	returnCode = http.StatusNotFound
	errMap = make(map[string]string)
	for key, err := range errorsMap {
		c := merry.RootCause(err)
		if c == nil {
			c = err
		}

		code := merry.HTTPCode(err)
		if code == http.StatusNotFound {
			continue
		} else if code == http.StatusInternalServerError && merry.Is(c, parser.ErrInvalidArg) {
			// check for invalid args, see applyByNode rewrite function
			code = http.StatusBadRequest
		}

		msg := merryError(c)
		errMap[key] = msg
		returnCode = recalcCode(returnCode, code)
	}

	return
}

func HttpErrorByCode(err merry.Error) merry.Error {
	var returnErr merry.Error
	if err == nil {
		returnErr = types.ErrNoMetricsFetched
	} else {
		code := merry.HTTPCode(err)
		msg := stripHtmlTags(merry.Message(err), 0)
		if code == http.StatusForbidden {
			returnErr = types.ErrForbidden
			if len(msg) > 0 {
				// pass message to caller
				returnErr = returnErr.WithMessage(msg)
			}
		} else if code == http.StatusServiceUnavailable || code == http.StatusBadGateway || code == http.StatusGatewayTimeout {
			returnErr = types.ErrFailedToFetch.WithHTTPCode(code).WithMessage(msg)
		} else {
			returnErr = types.ErrFailed.WithHTTPCode(code).WithMessage(msg)
		}
	}

	return returnErr
}
