// +build gofuzz

package expr

func Fuzz(data []byte) int {
	s := string(data)
	_, rem, err := ParseExpr(s)
	if rem == "" && err == nil {
		return 0
	}
	return 1
}
