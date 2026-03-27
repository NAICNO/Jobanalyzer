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
	refill func(verbose bool) (time.Time, time.Time, error)
	sync.Mutex
	timestamp time.Time
	minTime   time.Time
	maxTime   time.Time
}

func NewTimeCache(refill func(bool) (time.Time, time.Time, error)) *TimeCache {
	return &TimeCache{refill: refill}
}

func (tc *TimeCache) MinTime(soft, verbose bool) (time.Time, error) {
	tc.Lock()
	defer tc.Unlock()

	err := tc.maybeRefillLocked(soft, verbose)
	if err != nil {
		return time.Time{}, err
	}
	return tc.minTime, nil
}

func (tc *TimeCache) MaxTime(soft, verbose bool) (time.Time, error) {
	tc.Lock()
	defer tc.Unlock()

	err := tc.maybeRefillLocked(soft, verbose)
	if err != nil {
		return time.Time{}, err
	}
	return tc.maxTime, nil
}

func (tc *TimeCache) maybeRefillLocked(soft, verbose bool) error {
	now := time.Now()
	if !soft || tc.timestamp.Add(timeCacheLifetime).Before(now) {
		var reason string
		if tc.timestamp.IsZero() {
			reason = "empty"
		} else if !soft {
			reason = "hard"
		} else {
			reason = "expired"
		}
		newLow, newHigh, err := tc.refill(verbose)
		if err != nil {
			return err
		}
		tc.timestamp = now
		tc.minTime = newLow
		tc.maxTime = newHigh
		if verbose {
			Log.Infof("Refill min/max time cache bc %s: %v %v", reason, newLow, newHigh)
		}
	}
	return nil
}
