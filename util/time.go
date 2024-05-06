package util

import (
	"sync"
	"time"
)

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

type _locations struct {
	m    map[string]*time.Location
	lock sync.RWMutex
}

func (l *_locations) Get(name string) *time.Location {
	l.lock.RLock()
	defer l.lock.RUnlock()

	if loc := l.m[name]; loc != nil {
		return loc
	}

	l.lock.Lock()
	defer l.lock.Unlock()

	loc, err := time.LoadLocation(name)
	if err != nil {
		return nil
	}

	l.m[name] = loc
	return l.m[name]
}

var Locations = _locations{
	m:    make(map[string]*time.Location),
	lock: sync.RWMutex{},
}

func NowGMT8() time.Time {
	now := time.Now()

	loc := Locations.Get("Asia/Singapore")

	return now.In(loc)
}
