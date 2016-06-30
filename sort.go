package main

import (
	"sort"
	"strings"
)

// type for sorting a list of metrics by the nth part of the metric name.
// Implements sort.Interface minus Less, which needs to be provided by a struct
// that embeds this one. Provides compareBy for the benefit of that struct, which
// turns a function that compares two strings into a suitable Less function. Caches
// the relevant metric name part to avoid excessive calls to strings.Split.
type byPartBase struct {
	// the metrics to be sorted
	metrics []*metricData
	// which part of the name we are sorting on
	part int
	// a cache of the relevant part of the name for each metric in metrics
	keys []*string
}

func (b byPartBase) Len() int { return len(b.metrics) }

func (b byPartBase) Swap(i, j int) {
	b.metrics[i], b.metrics[j] = b.metrics[j], b.metrics[i]
	b.keys[i], b.keys[j] = b.keys[j], b.keys[i]
}

func getPart(metric *metricData, part int) *string {
	parts := strings.Split(metric.GetName(), ".")
	return &parts[part]
}

// Given two indices, i and j, and a comparator function that returns whether
// one metric name segment should sort before another, extracts the 'part'th part
// of the metric names, consults the comparator function, and returns a boolean
// suitable for use as the Less() method of a sort.Interface.
func (b byPartBase) compareBy(i, j int, comparator func(string, string) bool) bool {
	if b.keys[i] == nil {
		b.keys[i] = getPart(b.metrics[i], b.part)
	}
	if b.keys[j] == nil {
		b.keys[j] = getPart(b.metrics[j], b.part)
	}
	return comparator(*b.keys[i], *b.keys[j])
}

// Returns a byPartBase suitable for sorting 'metrics' by 'part'.
func ByPart(metrics []*metricData, part int) byPartBase {
	return byPartBase{
		metrics: metrics,
		keys:    make([]*string, len(metrics)),
		part:    part,
	}
}

// type for sorting a list of metrics 'alphabetically' (go string compare order)
type byPartAlphabetical struct {
	byPartBase
}

func (b byPartAlphabetical) Less(i, j int) bool {
	return b.compareBy(i, j, func(first, second string) bool {
		return first < second
	})
}

// returns a byPartAlphabetical that will sort 'metrics' alphabetically by 'part'.
func AlphabeticallyByPart(metrics []*metricData, part int) sort.Interface {
	return byPartAlphabetical{ByPart(metrics, part)}
}

// type for sorting a list of metrics 'by example' (a la perl Sort::ByExample)
// strings in the examples list sort in the order they appear in that list, while
// any strings not in the list sort at the end.
type byPartExample struct {
	byPartBase
	order map[string]int
}

func (b byPartExample) Less(i, j int) bool {
	return b.compareBy(i, j, func(first, second string) bool {
		return b.order[first] < b.order[second]
	})
}

// returns a byPartExample that will sort 'metrics', using the 'part'th part, with
// 'examples' as a list of examples to sort by.
func ByExample(metrics []*metricData, part int, examples []string) sort.Interface {
	order := map[string]int{}
	for i, example := range examples {
		// make them range from -n through -1, so that 0 (not found) will be last
		order[example] = i - len(examples)
	}

	return byPartExample{
		byPartBase: ByPart(metrics, part),
		order:      order,
	}
}

func sortMetrics(metrics []*metricData, mfetch metricRequest) {
	// Don't do any work if there are no globs in the metric name
	if !strings.ContainsAny(mfetch.metric, "*?[{") {
		return
	}
	parts := strings.Split(mfetch.metric, ".")
	// Proceed backwards by segments, sorting once for each segment that has a glob that calls for sorting.
	// By using a stable sort, the rightmost segments will be preserved as "sub-sorts" of any more leftward segments.
	for i := len(parts) - 1; i >= 0; i-- {
		if strings.ContainsAny(parts[i], "*?[") {
			sort.Stable(AlphabeticallyByPart(metrics, i))
		}
	}
}
