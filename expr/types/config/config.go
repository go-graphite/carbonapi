package config

type ConfigType = struct {
	// NudgeStartTimeOnAggregation enables nudging the start time of metrics
	// when aggregated. The start time is nudged in such way that timestamps
	// always fall in the same bucket. This is done by GraphiteWeb, and is
	// useful to avoid jitter in graphs when refreshing the page.
	NudgeStartTimeOnAggregation bool

	// UseBucketsHighestTimestampOnAggregation enables using the highest timestamp of the
	// buckets when aggregating to honor MaxDataPoints, instead of the lowest timestamp.
	// This prevents results to appear to predict the future.
	UseBucketsHighestTimestampOnAggregation bool
}

var Config = ConfigType{}
