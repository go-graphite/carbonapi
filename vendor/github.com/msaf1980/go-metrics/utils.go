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

func searchInt64Ge(a []int64, v int64) int {
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

func searchUint64Ge(a []uint64, v uint64) int {
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

func searchFloat64Ge(a []float64, v float64) int {
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
