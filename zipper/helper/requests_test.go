package helper

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

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

func Test_recalcCode(t *testing.T) {
	tests := []struct {
		code    int
		newCode int
		want    int
	}{
		{code: 500, newCode: 403, want: 403},
		{code: 403, newCode: 500, want: 403},
		{code: 403, newCode: 400, want: 400},
		{code: 400, newCode: 403, want: 400},
		{code: 500, newCode: 503, want: 500},
		{code: 503, newCode: 500, want: 500},
		{code: 503, newCode: 502, want: 503},
		{code: 0, newCode: 502, want: 503},
	}
	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			if got := recalcCode(tt.code, tt.newCode); got != tt.want {
				t.Errorf("recalcCode(%d, %d) = %d, want %d", tt.code, tt.newCode, got, tt.want)
			}
		})
	}
}
