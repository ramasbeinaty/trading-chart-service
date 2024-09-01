package utils

import "time"

func ConvertUnixMillisToTime(unixMillis int64) time.Time {
	seconds := unixMillis / 1000             // convert milliseconds to seconds
	nanoseconds := (unixMillis % 1000) * 1e6 // remainder as nanoseconds
	return time.Unix(seconds, nanoseconds).UTC()
}
