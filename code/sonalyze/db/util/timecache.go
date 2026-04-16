package util

import (
	"sync"
	"time"

	. "sonalyze/common"
)

const (
	timeCacheLifetime = 30 * time.Second
)

type TimeCache struct {
	refill func() (time.Time, time.Time, error)
	sync.Mutex
	timestamp time.Time
	minTime   time.Time
	maxTime   time.Time
}

func NewTimeCache(refill func() (time.Time, time.Time, error)) *TimeCache {
	return &TimeCache{refill: refill}
}

func (tc *TimeCache) MinTime(soft bool) (time.Time, error) {
	tc.Lock()
	defer tc.Unlock()

	err := tc.maybeRefillLocked(soft)
	if err != nil {
		return time.Time{}, err
	}
	return tc.minTime, nil
}

func (tc *TimeCache) MaxTime(soft bool) (time.Time, error) {
	tc.Lock()
	defer tc.Unlock()

	err := tc.maybeRefillLocked(soft)
	if err != nil {
		return time.Time{}, err
	}
	return tc.maxTime, nil
}

func (tc *TimeCache) maybeRefillLocked(soft bool) error {
	now := time.Now()
	if !soft || tc.timestamp.Add(timeCacheLifetime).Before(now) {
		newLow, newHigh, err := tc.refill()
		if err != nil {
			return err
		}
		tc.timestamp = now
		tc.minTime = newLow
		tc.maxTime = newHigh
		if Verbose {
			Log.Infof("Refill min/max time cache: %v %v", newLow, newHigh)
		}
	}
	return nil
}
