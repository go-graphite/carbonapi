package helper

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/go-graphite/carbonapi/expr/types"
)

// Total (sortByTotal), max (sortByMaxima), min (sortByMinima) sorting
// For 'min', we actually store 1/v so the sorting logic is the same
type ByVals struct {
	Vals   []float64
	Series []*types.MetricData
}

func (s ByVals) Len() int { return len(s.Series) }
func (s ByVals) Swap(i, j int) {
	s.Series[i], s.Series[j] = s.Series[j], s.Series[i]
	s.Vals[i], s.Vals[j] = s.Vals[j], s.Vals[i]
}
func (s ByVals) Less(i, j int) bool {
	// actually "greater than"
	return s.Vals[i] > s.Vals[j]
}

// ByName sorts metrics by name
type ByName []*types.MetricData

func (s ByName) Len() int           { return len(s) }
func (s ByName) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s ByName) Less(i, j int) bool { return s[i].Name < s[j].Name }

// ByNameNatural sorts metric naturally by name
type ByNameNatural []*types.MetricData

var dre = regexp.MustCompile(`\d+`)

func (s ByNameNatural) pad(str string) string {
	f := func(match []byte) []byte {
		n, _ := strconv.ParseInt(string(match), 10, 64)
		return []byte(fmt.Sprintf("%010d", n))
	}

	return string(dre.ReplaceAllFunc([]byte(str), f))
}
func (s ByNameNatural) Len() int           { return len(s) }
func (s ByNameNatural) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s ByNameNatural) Less(i, j int) bool { return s.pad(s[i].Name) < s.pad(s[j].Name) }
