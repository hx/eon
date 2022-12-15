package midground

import "time"

// schedule is used by Process to track its schedule.
type schedule struct {
	delay       time.Duration
	repeatAfter time.Duration
}
