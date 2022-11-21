package helper

import (
	"fmt"
	"net"
	"net/http"
	"reflect"
	"testing"

	"github.com/ansel1/merry"
	"github.com/go-graphite/carbonapi/zipper/types"
	"github.com/stretchr/testify/assert"
)

func TestMergeHttpErrors(t *testing.T) {
	type args struct {
	}
	tests := []struct {
		name     string
		errors   []merry.Error
		wantCode int
		want     []string
	}{
		{
			name:     "NotFound",
			errors:   []merry.Error{},
			wantCode: http.StatusNotFound,
			want:     []string{},
		},
		{
			name: "NetErr",
			errors: []merry.Error{
				types.ErrBackendError.WithValue("server", "test").WithCause(&net.OpError{Op: "connect", Err: fmt.Errorf("refused")}).WithHTTPCode(http.StatusServiceUnavailable),
			},
			wantCode: http.StatusServiceUnavailable,
			want:     []string{"connect: refused"},
		},
		{
			name: "NetErr (incapsulated)",
			errors: []merry.Error{
				types.ErrMaxTriesExceeded.WithCause(types.ErrBackendError.WithValue("server", "test").WithCause(&net.OpError{Op: "connect", Err: fmt.Errorf("refused")})).WithHTTPCode(http.StatusServiceUnavailable),
			},
			wantCode: http.StatusServiceUnavailable,
			want:     []string{"connect: refused"},
		},
		{
			name: "ServiceUnavailable",
			errors: []merry.Error{
				merry.New("unavaliable").WithHTTPCode(http.StatusServiceUnavailable),
			},
			wantCode: http.StatusServiceUnavailable,
			want:     []string{"unavaliable"},
		},
		{
			name: "GatewayTimeout and ServiceUnavailable",
			errors: []merry.Error{
				merry.New("timeout").WithHTTPCode(http.StatusGatewayTimeout),
				merry.New("unavaliable").WithHTTPCode(http.StatusServiceUnavailable),
			},
			wantCode: http.StatusServiceUnavailable,
			want:     []string{"timeout", "unavaliable"},
		},
		{
			name: "ServiceUnavailable and GatewayTimeout",
			errors: []merry.Error{
				merry.New("unavaliable").WithHTTPCode(http.StatusServiceUnavailable),
				merry.New("timeout").WithHTTPCode(http.StatusGatewayTimeout),
			},
			wantCode: http.StatusServiceUnavailable,
			want:     []string{"unavaliable", "timeout"},
		},
		{
			name: "Forbidden and GatewayTimeout",
			errors: []merry.Error{
				merry.New("limit").WithHTTPCode(http.StatusForbidden),
				merry.New("timeout").WithHTTPCode(http.StatusGatewayTimeout),
			},
			wantCode: http.StatusForbidden,
			want:     []string{"limit", "timeout"},
		},
		{
			name: "GatewayTimeout and Forbidden",
			errors: []merry.Error{
				merry.New("timeout").WithHTTPCode(http.StatusGatewayTimeout),
				merry.New("limit").WithHTTPCode(http.StatusForbidden),
			},
			wantCode: http.StatusForbidden,
			want:     []string{"timeout", "limit"},
		},
		{
			name: "InternalServerError and Forbidden",
			errors: []merry.Error{
				merry.New("error").WithHTTPCode(http.StatusInternalServerError),
				merry.New("limit").WithHTTPCode(http.StatusForbidden),
			},
			wantCode: http.StatusForbidden,
			want:     []string{"error", "limit"},
		},
		{
			name: "InternalServerError and GatewayTimeout",
			errors: []merry.Error{
				merry.New("error").WithHTTPCode(http.StatusInternalServerError),
				merry.New("timeout").WithHTTPCode(http.StatusGatewayTimeout),
			},
			wantCode: http.StatusInternalServerError,
			want:     []string{"error", "timeout"},
		},
		{
			name: "GatewayTimeout and InternalServerError",
			errors: []merry.Error{
				merry.New("timeout").WithHTTPCode(http.StatusGatewayTimeout),
				merry.New("error").WithHTTPCode(http.StatusInternalServerError),
			},
			wantCode: http.StatusInternalServerError,
			want:     []string{"timeout", "error"},
		},
		{
			name: "BadRequest and Forbidden",
			errors: []merry.Error{
				merry.New("error").WithHTTPCode(http.StatusBadRequest),
				merry.New("limit").WithHTTPCode(http.StatusForbidden),
			},
			wantCode: http.StatusForbidden, // Last win
			want:     []string{"error", "limit"},
		},
		{
			name: "Forbidden and BadRequest",
			errors: []merry.Error{
				merry.New("limit").WithHTTPCode(http.StatusForbidden),
				merry.New("error").WithHTTPCode(http.StatusBadRequest),
			},
			wantCode: http.StatusBadRequest, // Last win
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

func Test_stripHtmlTags(t *testing.T) {
	tests := []struct {
		name   string
		s      string
		maxLen int
		want   string
	}{
		{
			name:   "Empty",
			s:      "",
			maxLen: 10,
			want:   "",
		},
		{
			name:   "Broken #1",
			s:      "<html>\r\n<head",
			maxLen: 0,
			want:   "head",
		},
		{
			name:   "Broken #2",
			s:      "<html>\r\nhead>",
			maxLen: 0,
			want:   "head>",
		},
		{
			name:   "Nginx Gateway Timeout",
			s:      "<html>\r\n<head><title>504 Gateway Time-out</title></head>\r\n<body>\r\n<center><h1>504 Gateway Time-out</h1></center>\r\n<hr><center>nginx</center>\r\n</body>\r\n</html>\r",
			maxLen: 0,
			want:   "504 Gateway Time-out\r\n\r\n504 Gateway Time-out\r\nnginx",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, stripHtmlTags(tt.s, tt.maxLen))
		})
	}
}
