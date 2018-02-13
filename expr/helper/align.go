package helper

import (
	"math"
	"time"
)

func GetBuckets(start, stop, bucketSize int32) int32 {
	return int32(math.Ceil(float64(stop-start) / float64(bucketSize)))
}

func AlignStartToInterval(start, stop, bucketSize int32) int32 {
	for _, v := range []int32{86400, 3600, 60} {
		if bucketSize >= v {
			start -= start % v
			break
		}
	}

	return start
}

func AlignToBucketSize(start, stop, bucketSize int32) (int32, int32) {
	start = int32(time.Unix(int64(start), 0).Truncate(time.Duration(bucketSize) * time.Second).Unix())
	newStop := int32(time.Unix(int64(stop), 0).Truncate(time.Duration(bucketSize) * time.Second).Unix())

	// check if a partial bucket is needed
	if stop != newStop {
		newStop += bucketSize
	}

	return start, newStop
}
