package promotion

import (
	"testing"
	"time"
)

func TestTime(t *testing.T) {
	loc := time.FixedZone("", 28800)
	date := time.Now().In(loc)
	today := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, loc)
	duration := time.Until(today.Add(24 * time.Hour))
	t.Log(today)
	t.Log(duration)
	t.Error(1)
}
