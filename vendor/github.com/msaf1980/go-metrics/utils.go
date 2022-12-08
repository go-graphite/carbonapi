package metrics

func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func Min(a, b int) int {
	if a <= b {
		return a
	}
	return b
}

func IsSortedSliceInt64Ge(a []int64) (sorted bool) {
	if len(a) == 0 {
		return
	}
	sorted = true
	prev := a[0]
	for i := 1; i < len(a); i++ {
		if a[i] <= prev {
			sorted = false
			break
		}
		prev = a[i]
	}
	return
}

func IsSortedSliceUint64Le(a []uint64) (sorted bool) {
	if len(a) == 0 {
		return
	}
	sorted = true
	prev := a[0]
	for i := 1; i < len(a); i++ {
		if a[i] <= prev {
			sorted = false
			break
		}
		prev = a[i]
	}
	return
}

func IsSortedSliceFloat64Le(a []float64) (sorted bool) {
	if len(a) == 0 {
		return
	}
	sorted = true
	prev := a[0]
	for i := 1; i < len(a); i++ {
		if a[i] <= prev {
			sorted = false
			break
		}
		prev = a[i]
	}
	return
}

func SearchInt64Le(a []int64, v int64) int {
	if v <= a[0] {
		return 0
	}
	start := 0
	end := len(a) - 1
	// if end == 0 || v > a[end] {
	// 	return end
	// }
	for end > start {
		mid := start + (end-start)/2
		if v > a[mid] {
			start = mid + 1
		} else if v == a[mid] {
			return mid
		} else {
			end = mid
		}
	}
	return end
}

func SearchUint64Le(a []uint64, v uint64) int {
	if v <= a[0] {
		return 0
	}
	start := 0
	end := len(a) - 1
	// if end == 0 || v > a[end] {
	// 	return end
	// }
	for end > start {
		mid := start + (end-start)/2
		if v > a[mid] {
			start = mid + 1
		} else if v == a[mid] {
			return mid
		} else {
			end = mid
		}
	}
	return end
}

func SearchFloat64Le(a []float64, v float64) int {
	if v <= a[0] {
		return 0
	}
	start := 0
	end := len(a) - 1
	if end == 0 || v > a[end] {
		return end
	}
	for end > start {
		mid := start + (end-start)/2
		if v > a[mid] {
			start = mid + 1
		} else if v == a[mid] {
			return mid
		} else {
			end = mid
		}
	}
	return end
}
