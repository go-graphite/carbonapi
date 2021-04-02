package http

import (
	"net"
	"net/http"
)

type HttpError struct {
	Method   string
	Query    string
	Err      error
	HttpCode int
}

func HttpCode(err error, found bool) int {
	if err == nil {
		if !found {
			return http.StatusNotFound
		}
		return http.StatusOK
	}
	if _, ok := err.(*net.OpError); ok {
		return http.StatusServiceUnavailable
	}
	return http.StatusInternalServerError
}

func NewHttpError(method string, query string, err error) *HttpError {
	return &HttpError{
		Method:   method,
		Query:    query,
		Err:      err,
		HttpCode: HttpCode(err, false),
	}
}

func NewHttpNotFound(method string, query string) *HttpError {
	return &HttpError{
		Method:   method,
		Query:    query,
		Err:      nil,
		HttpCode: http.StatusNotFound,
	}
}

func (e *HttpError) Error() string {
	return e.Method + " " + e.Query + ": " + e.Err.Error()
}
