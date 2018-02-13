package types

type MetricHeapElement struct {
	Idx int
	Val float64
}

type MetricHeap []MetricHeapElement

func (m MetricHeap) Len() int           { return len(m) }
func (m MetricHeap) Swap(i, j int)      { m[i], m[j] = m[j], m[i] }
func (m MetricHeap) Less(i, j int) bool { return m[i].Val < m[j].Val }

func (m *MetricHeap) Push(x interface{}) {
	*m = append(*m, x.(MetricHeapElement))
}

func (m *MetricHeap) Pop() interface{} {
	old := *m
	n := len(old)
	x := old[n-1]
	*m = old[0 : n-1]
	return x
}
