package helper

import (
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
			name:     "404 (empty)",
			errors:   []merry.Error{},
			wantCode: 404,
			want:     []string{},
		},
		{
			name: "503",
			errors: []merry.Error{
				merry.New("unavaliable").WithHTTPCode(503),
			},
			wantCode: 503,
			want:     []string{"unavaliable"},
		},
		{
			name: "504 and 503",
			errors: []merry.Error{
				merry.New("timeout").WithHTTPCode(504),
				merry.New("unavaliable").WithHTTPCode(503),
			},
			wantCode: 503,
			want:     []string{"timeout", "unavaliable"},
		},
		{
			name: "503 and 504",
			errors: []merry.Error{
				merry.New("unavaliable").WithHTTPCode(503),
				merry.New("timeout").WithHTTPCode(504),
			},
			wantCode: 503,
			want:     []string{"unavaliable", "timeout"},
		},
		{
			name: "403 and 504",
			errors: []merry.Error{
				merry.New("limit").WithHTTPCode(403),
				merry.New("timeout").WithHTTPCode(504),
			},
			wantCode: 403,
			want:     []string{"limit", "timeout"},
		},
		{
			name: "504 and 403",
			errors: []merry.Error{
				merry.New("timeout").WithHTTPCode(504),
				merry.New("limit").WithHTTPCode(403),
			},
			wantCode: 403,
			want:     []string{"timeout", "limit"},
		},
		{
			name: "500 and 403",
			errors: []merry.Error{
				merry.New("error").WithHTTPCode(500),
				merry.New("limit").WithHTTPCode(403),
			},
			wantCode: 403,
			want:     []string{"error", "limit"},
		},
		{
			name: "500 and 504",
			errors: []merry.Error{
				merry.New("error").WithHTTPCode(500),
				merry.New("timeout").WithHTTPCode(504),
			},
			wantCode: 500,
			want:     []string{"error", "timeout"},
		},
		{
			name: "504 and 500",
			errors: []merry.Error{
				merry.New("timeout").WithHTTPCode(504),
				merry.New("error").WithHTTPCode(500),
			},
			wantCode: 500,
			want:     []string{"timeout", "error"},
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
