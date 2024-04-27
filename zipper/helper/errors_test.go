package helper

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"reflect"
	"testing"

	"github.com/ansel1/merry/v2"

	"github.com/go-graphite/carbonapi/zipper/types"
)

func TestMergeHttpErrors(t *testing.T) {
	tests := []struct {
		name     string
		errors   []error
		wantCode int
		want     []string
	}{
		{
			name:     "NotFound",
			errors:   []error{},
			wantCode: http.StatusNotFound,
			want:     []string{},
		},
		{
			name: "NetErr",
			errors: []error{
				types.ErrBackendError.WithValue("server", "test").WithCause(&net.OpError{Op: "connect", Err: fmt.Errorf("refused")}).WithHTTPCode(http.StatusServiceUnavailable),
			},
			wantCode: http.StatusServiceUnavailable,
			want:     []string{"connect: refused"},
		},
		{
			name: "NetErr (incapsulated)",
			errors: []error{
				types.ErrMaxTriesExceeded.WithCause(types.ErrBackendError.WithValue("server", "test").WithCause(&net.OpError{Op: "connect", Err: fmt.Errorf("refused")})).WithHTTPCode(http.StatusServiceUnavailable),
			},
			wantCode: http.StatusServiceUnavailable,
			want:     []string{"connect: refused"},
		},
		{
			name: "ServiceUnavailable",
			errors: []error{
				merry.Wrap(errors.New("unavaliable"), merry.WithHTTPCode(http.StatusServiceUnavailable)),
			},
			wantCode: http.StatusServiceUnavailable,
			want:     []string{"unavaliable"},
		},
		{
			name: "GatewayTimeout and ServiceUnavailable",
			errors: []error{
				merry.Wrap(errors.New("timeout"), merry.WithHTTPCode(http.StatusGatewayTimeout)),
				merry.Wrap(errors.New("unavaliable"), merry.WithHTTPCode(http.StatusServiceUnavailable)),
			},
			wantCode: http.StatusServiceUnavailable,
			want:     []string{"timeout", "unavaliable"},
		},
		{
			name: "ServiceUnavailable and GatewayTimeout",
			errors: []error{
				merry.Wrap(errors.New("unavaliable"), merry.WithHTTPCode(http.StatusServiceUnavailable)),
				merry.Wrap(errors.New("timeout"), merry.WithHTTPCode(http.StatusGatewayTimeout)),
			},
			wantCode: http.StatusServiceUnavailable,
			want:     []string{"unavaliable", "timeout"},
		},
		{
			name: "Forbidden and GatewayTimeout",
			errors: []error{
				merry.Wrap(errors.New("limit"), merry.WithHTTPCode(http.StatusForbidden)),
				merry.Wrap(errors.New("timeout"), merry.WithHTTPCode(http.StatusGatewayTimeout)),
			},
			wantCode: http.StatusForbidden,
			want:     []string{"limit", "timeout"},
		},
		{
			name: "GatewayTimeout and Forbidden",
			errors: []error{
				merry.Wrap(errors.New("timeout"), merry.WithHTTPCode(http.StatusGatewayTimeout)),
				merry.Wrap(errors.New("limit"), merry.WithHTTPCode(http.StatusForbidden)),
			},
			wantCode: http.StatusForbidden,
			want:     []string{"timeout", "limit"},
		},
		{
			name: "InternalServerError and Forbidden",
			errors: []error{
				merry.Wrap(errors.New("error"), merry.WithHTTPCode(http.StatusInternalServerError)),
				merry.Wrap(errors.New("limit"), merry.WithHTTPCode(http.StatusForbidden)),
			},
			wantCode: http.StatusForbidden,
			want:     []string{"error", "limit"},
		},
		{
			name: "InternalServerError and GatewayTimeout",
			errors: []error{
				merry.Wrap(errors.New("error"), merry.WithHTTPCode(http.StatusInternalServerError)),
				merry.Wrap(errors.New("timeout"), merry.WithHTTPCode(http.StatusGatewayTimeout)),
			},
			wantCode: http.StatusInternalServerError,
			want:     []string{"error", "timeout"},
		},
		{
			name: "GatewayTimeout and InternalServerError",
			errors: []error{
				merry.Wrap(errors.New("timeout"), merry.WithHTTPCode(http.StatusGatewayTimeout)),
				merry.Wrap(errors.New("error"), merry.WithHTTPCode(http.StatusInternalServerError)),
			},
			wantCode: http.StatusInternalServerError,
			want:     []string{"timeout", "error"},
		},
		{
			name: "BadRequest and Forbidden",
			errors: []error{
				merry.Wrap(errors.New("error"), merry.WithHTTPCode(http.StatusBadRequest)),
				merry.Wrap(errors.New("limit"), merry.WithHTTPCode(http.StatusForbidden)),
			},
			wantCode: http.StatusBadRequest,
			want:     []string{"error", "limit"},
		},
		{
			name: "Forbidden and BadRequest",
			errors: []error{
				merry.Wrap(errors.New("limit"), merry.WithHTTPCode(http.StatusForbidden)),
				merry.Wrap(errors.New("error"), merry.WithHTTPCode(http.StatusBadRequest)),
			},
			wantCode: http.StatusBadRequest,
			want:     []string{"limit", "error"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCode, got := MergeHttpErrors(tt.errors)
			if gotCode != tt.wantCode {
				t.Errorf("MergeHttpErrors() gotCode = %v, want %v", gotCode, tt.wantCode)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MergeHttpErrors() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMergeHttpErrorMap(t *testing.T) {
	tests := []struct {
		name     string
		errors   map[string]error
		wantCode int
		want     map[string]string
	}{
		{
			name:     "NotFound",
			errors:   map[string]error{},
			wantCode: http.StatusNotFound,
			want:     map[string]string{},
		},
		{
			name: "NetErr",
			errors: map[string]error{
				"a": types.ErrBackendError.WithValue("server", "test").WithCause(&net.OpError{Op: "connect", Err: fmt.Errorf("refused")}).WithHTTPCode(http.StatusServiceUnavailable),
			},
			wantCode: http.StatusServiceUnavailable,
			want:     map[string]string{"a": "connect: refused"},
		},
		{
			name: "NetErr (incapsulated)",
			errors: map[string]error{
				"b": types.ErrMaxTriesExceeded.WithCause(types.ErrBackendError.WithValue("server", "test").WithCause(&net.OpError{Op: "connect", Err: fmt.Errorf("refused")})).WithHTTPCode(http.StatusServiceUnavailable),
			},
			wantCode: http.StatusServiceUnavailable,
			want:     map[string]string{"b": "connect: refused"},
		},
		{
			name: "ServiceUnavailable",
			errors: map[string]error{
				"d": merry.Wrap(errors.New("unavaliable"), merry.WithHTTPCode(http.StatusServiceUnavailable)),
			},
			wantCode: http.StatusServiceUnavailable,
			want:     map[string]string{"d": "unavaliable"},
		},
		{
			name: "GatewayTimeout and ServiceUnavailable",
			errors: map[string]error{
				"a":  merry.Wrap(errors.New("timeout"), merry.WithHTTPCode(http.StatusGatewayTimeout)),
				"de": merry.Wrap(errors.New("unavaliable"), merry.WithHTTPCode(http.StatusServiceUnavailable)),
			},
			wantCode: http.StatusServiceUnavailable,
			want:     map[string]string{"a": "timeout", "de": "unavaliable"},
		},
		{
			name: "ServiceUnavailable and GatewayTimeout",
			errors: map[string]error{
				"de": merry.Wrap(errors.New("unavaliable"), merry.WithHTTPCode(http.StatusServiceUnavailable)),
				"a":  merry.Wrap(errors.New("timeout"), merry.WithHTTPCode(http.StatusGatewayTimeout)),
			},
			wantCode: http.StatusServiceUnavailable,
			want:     map[string]string{"a": "timeout", "de": "unavaliable"},
		},
		{
			name: "Forbidden and GatewayTimeout",
			errors: map[string]error{
				"de": merry.Wrap(errors.New("limit"), merry.WithHTTPCode(http.StatusForbidden)),
				"c":  merry.Wrap(errors.New("timeout"), merry.WithHTTPCode(http.StatusGatewayTimeout)),
			},
			wantCode: http.StatusForbidden,
			want:     map[string]string{"c": "timeout", "de": "limit"},
		},
		{
			name: "GatewayTimeout and Forbidden",
			errors: map[string]error{
				"a": merry.Wrap(errors.New("limit"), merry.WithHTTPCode(http.StatusForbidden)),
				"c": merry.Wrap(errors.New("timeout"), merry.WithHTTPCode(http.StatusGatewayTimeout)),
			},
			wantCode: http.StatusForbidden,
			want:     map[string]string{"a": "limit", "c": "timeout"},
		},
		{
			name: "InternalServerError and Forbidden",
			errors: map[string]error{
				"a":  merry.Wrap(errors.New("error"), merry.WithHTTPCode(http.StatusInternalServerError)),
				"cd": merry.Wrap(errors.New("limit"), merry.WithHTTPCode(http.StatusForbidden)),
			},
			wantCode: http.StatusForbidden,
			want:     map[string]string{"a": "error", "cd": "limit"},
		},
		{
			name: "InternalServerError and GatewayTimeout",
			errors: map[string]error{
				"a": merry.Wrap(errors.New("error"), merry.WithHTTPCode(http.StatusInternalServerError)),
				"b": merry.Wrap(errors.New("timeout"), merry.WithHTTPCode(http.StatusGatewayTimeout)),
			},
			wantCode: http.StatusInternalServerError,
			want:     map[string]string{"a": "error", "b": "timeout"},
		},
		{
			name: "GatewayTimeout and InternalServerError",
			errors: map[string]error{
				"a":  merry.Wrap(errors.New("timeout"), merry.WithHTTPCode(http.StatusGatewayTimeout)),
				"cd": merry.Wrap(errors.New("error"), merry.WithHTTPCode(http.StatusInternalServerError)),
			},
			wantCode: http.StatusInternalServerError,
			want:     map[string]string{"a": "timeout", "cd": "error"},
		},
		{
			name: "BadRequest and Forbidden",
			errors: map[string]error{
				"de": merry.Wrap(errors.New("error"), merry.WithHTTPCode(http.StatusBadRequest)),
				"a":  merry.Wrap(errors.New("limit"), merry.WithHTTPCode(http.StatusForbidden)),
			},
			wantCode: http.StatusBadRequest,
			want:     map[string]string{"a": "limit", "de": "error"},
		},
		{
			name: "Forbidden and BadRequest",
			errors: map[string]error{
				"a": merry.Wrap(errors.New("limit"), merry.WithHTTPCode(http.StatusForbidden)),
				"b{c,de,klmn}.cde.d{c,de,klmn}.e{c,de,klmn}.k{c,de,klmn}.b{c,de,klmn}.cde.d{c,de,klmn}.e{c,de,klmn}.k{c,de,klmn}.e{c,de,klmn}.k{c,de,klmn}": merry.Wrap(errors.New("error"), merry.WithHTTPCode(http.StatusBadRequest)),
			},
			wantCode: http.StatusBadRequest,
			want: map[string]string{
				"a": "limit",
				"b{c,de,klmn}.cde.d{c,de,klmn}.e{c,de,klmn}.k{c,de,klmn}.b{c,de,klmn}.cde.d{c,de,klmn}.e{c,de,klmn}.k{c,de,klmn}.e{c,de,klmn}.k{c,de,klmn}": "error",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCode, got := MergeHttpErrorMap(tt.errors)
			if gotCode != tt.wantCode {
				t.Errorf("MergeHttpErrors() gotCode = %v, want %v", gotCode, tt.wantCode)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MergeHttpErrors() got = %v, want %v", got, tt.want)
			}
		})
	}
}
