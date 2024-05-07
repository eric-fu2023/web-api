package util

import (
	"errors"
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

func (l *_locations) Get(name string) (*time.Location, error) {
	l.lock.RLock()
	loc := l.m[name]
	l.lock.RUnlock()

	if loc != nil {
		return loc, nil
	}

	l.lock.Lock()
	defer l.lock.Unlock()

	loc, err := time.LoadLocation(name)
	if err != nil {
		return nil, err
	}

	if loc == nil {
		return nil, errors.New("location not found")
	}

	l.m[name] = loc
	return l.m[name], nil
}

var Locations = _locations{
	m:    make(map[string]*time.Location),
	lock: sync.RWMutex{},
}

func NowGMT8() (time.Time, error) {
	now := time.Now()

	loc, err := Locations.Get("Asia/Singapore")
	if err != nil || loc == nil {
		return time.Time{}, errors.New("[NowGMT8] can't get timezone")
	}
	return now.In(loc), nil
}

func LastDayOfPreviousMonth(t time.Time) time.Time {
	y, m, _ := t.Date()
	return time.Date(y, m, 0, 0, 0, 0, 0, t.Location())
}
