package util

import "time"

// GetRemainSecondsToday returns the remaining seconds until end of current day (local timezone).
func GetRemainSecondsToday() int {
	now := time.Now()
	endOfDay := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location())
	return int(endOfDay.Sub(now).Seconds())
}

// GetStartOfDay returns the start of today (00:00:00).
func GetStartOfDay() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
}

// IsToday checks if the given time is today.
func IsToday(t time.Time) bool {
	now := time.Now()
	return t.Year() == now.Year() && t.Month() == now.Month() && t.Day() == now.Day()
}
