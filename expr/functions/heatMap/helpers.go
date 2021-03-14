package heatMap

import (
	"fmt"
	"math"
	"sort"

	"github.com/go-graphite/carbonapi/expr/types"
)

// Helper functions that are used to sort or validate metrics
func validateNeighbourSeries(series []*types.MetricData) error {
	if len(series) == 0 {
		return nil
	}
	s1 := series[0]

	for i := 1; i < len(series); i++ {
		s2 := series[i]
		if s1.StartTime != s2.StartTime {
			return fmt.Errorf("StartTime differs: %d!=%d", s1.StartTime, s2.StartTime)
		}
		if s1.StopTime != s2.StopTime {
			return fmt.Errorf("StartTime differs: %d!=%d", s1.StopTime, s2.StopTime)
		}
		if s1.StepTime != s2.StepTime {
			return fmt.Errorf("StartTime differs: %d!=%d", s1.StepTime, s2.StepTime)
		}
		if len(s1.Values) != len(s2.Values) {
			return fmt.Errorf("values quantity differs: %d!=%d", len(s1.Values), len(s2.Values))
		}
	}
	return nil
}

// sortMetricData returns *types.MetricData list sorted by sum of the first values
func sortMetricData(list []*types.MetricData) []*types.MetricData {
	// take 5 first not null values
	const points = 5

	// mate series with its weight (sum of first values)
	type metricDataWeighted struct {
		data   *types.MetricData
		weight float64
	}

	seriesQty := len(list)
	if seriesQty < 2 {
		return list
	}

	listWeighted := make([]metricDataWeighted, seriesQty)
	for j := 0; j < seriesQty; j++ {
		listWeighted[j].data = list[j]
	}

	pointsFound := 0
	valuesQty := len(list[0].Values)

	for i := 0; i < valuesQty && pointsFound < points; i++ {
		// make sure that each series has current point not null
		absent := false
		for j := 0; j < seriesQty && !absent; j++ {
			absent = math.IsNaN(list[j].Values[i])
		}
		if absent {
			continue
		}

		// accumulate sum of first not-null values
		for j := 0; j < seriesQty; j++ {
			listWeighted[j].weight += list[j].Values[i]
		}
		pointsFound++
	}

	// sort series by its weight
	if pointsFound > 0 {
		sort.SliceStable(listWeighted, func(i, j int) bool {
			return listWeighted[i].weight < listWeighted[j].weight
		})
		for j := 0; j < seriesQty; j++ {
			list[j] = listWeighted[j].data
		}
	}

	return list
}
