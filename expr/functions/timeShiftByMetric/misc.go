package timeShiftByMetric

import (
	"sort"
	"strconv"

	"github.com/ansel1/merry"

	"github.com/go-graphite/carbonapi/expr/types"
)

var (
	errSeriesLengthMismatch = merry.Errorf("bad data: length of Values for series differs from others")
	errTooFewDatasets = merry.Errorf("bad data: too few data sets")
	errLessThan2Marks = merry.Errorf("bad data: could not find 2 marks")
	errEmptySeries = merry.Errorf("bad data: empty series")
)

type callParams struct {
	marks       []*types.MetricData
	metrics     []*types.MetricData
	versionRank int
	pointsQty   int
	stepTime    int64
}

type versionInfo struct {
	mark         string
	position     int
	versionMajor int
	versionMinor int
}

type versionInfos []versionInfo

// HighestVersions returns slice of markVersionInfo
// containing the highest version for each major version
func (data versionInfos) HighestVersions() versionInfos {
	qty := 0
	result := make(versionInfos, 0, len(data))

	sort.Sort(sort.Reverse(data))
	for _, current := range data {
		if qty == 0 || result[qty-1].versionMajor != current.versionMajor {
			result = append(result, current)
			qty++
		}
	}

	return result
}

func (data versionInfos) Len() int {
	return len(data)
}

func (data versionInfos) Less(i, j int) bool {
	if data[i].versionMajor == data[j].versionMajor {
		return data[i].versionMinor < data[j].versionMinor
	} else {
		return data[i].versionMajor < data[j].versionMajor
	}
}

func (data versionInfos) Swap(i, j int) {
	data[i], data[j] = data[j], data[i]
}

// mustAtoi is like strconv.Atoi, but causes panic in case of error
func mustAtoi(s string) int {
	if result, err := strconv.Atoi(s); err != nil {
		panic(err)
	} else {
		return result
	}
}
