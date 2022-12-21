package events

import "github.com/hx/eon"

// MultiDelegate is a midground.Delegate implementation that re-sends notifications to multiple delegates.
type MultiDelegate []eon.Delegate

func (m MultiDelegate) JobScheduled(process *eon.Process) {
	for _, d := range m {
		if d != nil {
			d.JobScheduled(process)
		}
	}
}

func (m MultiDelegate) JobBlocked(process *eon.Process, blockers []*eon.Process) {
	for _, d := range m {
		if d != nil {
			d.JobBlocked(process, blockers)
		}
	}
}

func (m MultiDelegate) JobStarting(process *eon.Process) {
	for _, d := range m {
		if d != nil {
			d.JobStarting(process)
		}
	}
}

func (m MultiDelegate) JobProgressed(process *eon.Process, payload any) {
	for _, d := range m {
		if d != nil {
			d.JobProgressed(process, payload)
		}
	}
}

func (m MultiDelegate) JobEnded(process *eon.Process, err error) {
	for _, d := range m {
		if d != nil {
			d.JobEnded(process, err)
		}
	}
}
