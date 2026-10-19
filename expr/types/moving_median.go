package types

import (
	"container/heap"
	"math"
)

type medianSlot struct {
	value float64
	heap  *medianHeap
	index int
}

type medianHeap struct {
	slots []*medianSlot
	max   bool
}

func (h *medianHeap) Len() int { return len(h.slots) }

func (h *medianHeap) Less(i, j int) bool {
	if h.max {
		return h.slots[i].value > h.slots[j].value
	}
	return h.slots[i].value < h.slots[j].value
}

func (h *medianHeap) Swap(i, j int) {
	h.slots[i], h.slots[j] = h.slots[j], h.slots[i]
	h.slots[i].index = i
	h.slots[j].index = j
}

func (h *medianHeap) Push(x any) {
	s := x.(*medianSlot)
	s.heap = h
	s.index = len(h.slots)
	h.slots = append(h.slots, s)
}

func (h *medianHeap) Pop() any {
	last := len(h.slots) - 1
	s := h.slots[last]
	h.slots = h.slots[:last]
	s.heap = nil
	return s
}

func (h *medianHeap) top() float64 { return h.slots[0].value }

// movingMedian keeps the non-NaN values of the window in two heaps: lower (max-heap) holds the
// smaller half and upper (min-heap) the larger half, so the median is read from their tops.
// Each slot tracks its heap position, so a value that leaves the window is removed in O(log n).
type movingMedian struct {
	slots []medianSlot
	lower medianHeap
	upper medianHeap
}

func newMovingMedian(data []float64) *movingMedian {
	m := &movingMedian{
		slots: make([]medianSlot, len(data)),
		lower: medianHeap{slots: make([]*medianSlot, 0, len(data)), max: true},
		upper: medianHeap{slots: make([]*medianSlot, 0, len(data))},
	}
	for i, v := range data {
		m.replace(i, v)
	}
	return m
}

func (m *movingMedian) replace(pos int, v float64) {
	s := &m.slots[pos]
	if s.heap != nil {
		heap.Remove(s.heap, s.index)
		m.rebalance()
	}

	s.value = v
	if math.IsNaN(v) {
		return
	}
	if m.lower.Len() == 0 || v <= m.lower.top() {
		heap.Push(&m.lower, s)
	} else {
		heap.Push(&m.upper, s)
	}
	m.rebalance()
}

func (m *movingMedian) rebalance() {
	switch {
	case m.lower.Len() > m.upper.Len()+1:
		heap.Push(&m.upper, heap.Pop(&m.lower))
	case m.upper.Len() > m.lower.Len():
		heap.Push(&m.lower, heap.Pop(&m.upper))
	}
}

func (m *movingMedian) median() float64 {
	switch {
	case m.lower.Len() == 0:
		return math.NaN()
	case m.lower.Len() > m.upper.Len():
		return m.lower.top()
	default:
		return (m.lower.top() + m.upper.top()) / 2
	}
}
