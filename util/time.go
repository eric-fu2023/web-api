package util

import "time"

// RoundDownTimeDay rounds down the given time to the nearest day.
// time.Truncate cant be used on day durations as Truncate assumes UTC timezone
// and returns erroneous results with durations >= 1 day
func RoundDownTimeDay(t time.Time) time.Time {
	r := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	return r
}

// RoundUpTimeDay rounds up the given time to the nearest day.
// time.Truncate cant be used on day durations as Truncate assumes UTC timezone
// and returns erroneous results with durations >= 1 day
func RoundUpTimeDay(t time.Time) time.Time {
	t = t.Add(-time.Nanosecond)
	r := time.Date(t.Year(), t.Month(), t.Day()+1, 0, 0, 0, 0, t.Location())
	return r
}
