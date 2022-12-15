package midground

import "time"

type clock interface {
	Now() time.Time
	After(d time.Duration) <-chan time.Time
}

type systemClock struct{}

func (s systemClock) Now() time.Time                         { return time.Now() }
func (s systemClock) After(d time.Duration) <-chan time.Time { return time.After(d) }
