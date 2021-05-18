package helper

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/ansel1/merry"
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
