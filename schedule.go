package midground

import "time"

type Schedule struct {
	Delay       time.Duration
	RepeatAfter time.Duration
}
