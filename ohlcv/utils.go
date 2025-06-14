package ohlcv

import "time"

// getBucketTime input tradeTime and interval return the timestamp align the interval boundary
// Example: 1 hr boundary: (1)2024-01-01 00:00:00, (2)2024-01-01 00:01:00, (3)2024-01-01 00:02:00 (4)...
func getBucketUnixTime(tradeTime time.Time, interval time.Duration) int64 {
	openTime := tradeTime.Truncate(interval).Unix()
	return openTime
}

func getBucketTime(tradeTime time.Time, interval time.Duration) time.Time {
	return tradeTime.Truncate(interval)
}

// getNextBucketTime
func getNextBucketTime(current time.Time, interval time.Duration) time.Time {
	next := current.Add(interval)
	return next.Truncate(interval)
}

// getNextBucketTime
func getNextBucketUnixTime(current time.Time, interval time.Duration) int64 {
	next := current.Add(interval)
	return next.Truncate(interval).Unix()
}
