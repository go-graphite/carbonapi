package helper

import (
	"math"
	"time"
)

// GetBuckets returns amount buckets for timeSeries (defined with startTime, stopTime and step (bucket) size.
func GetBuckets(start, stop, bucketSize uint32) uint32 {
	return uint32(math.Ceil(float64(stop-start) / float64(bucketSize)))
}

// AlignStartToInterval aligns start of serie to interval
func AlignStartToInterval(start, stop, bucketSize uint32) uint32 {
	for _, v := range []uint32{86400, 3600, 60} {
		if bucketSize >= v {
			start -= start % v
			break
		}
	}

	return start
}

// AlignToBucketSize aligns start and stop of serie to specified bucket (step) size
func AlignToBucketSize(start, stop, bucketSize uint32) (uint32, uint32) {
	start = uint32(time.Unix(int64(start), 0).Truncate(time.Duration(bucketSize) * time.Second).Unix())
	newStop := uint32(time.Unix(int64(stop), 0).Truncate(time.Duration(bucketSize) * time.Second).Unix())

	// check if a partial bucket is needed
	if stop != newStop {
		newStop += bucketSize
	}

	return start, newStop
}
