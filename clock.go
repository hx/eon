package eon

import "time"

// clock is used by Scheduler as a stand-in for the system clock. It's a placeholder for if/when we want to allow the
// clock to be stubbed for testing Scheduler.
type clock interface {
	Now() time.Time
	After(d time.Duration) <-chan time.Time
}

type systemClock struct{}

// Now proxies time.Now.
func (s systemClock) Now() time.Time { return time.Now() }

// After proxies time.After.
func (s systemClock) After(d time.Duration) <-chan time.Time { return time.After(d) }
