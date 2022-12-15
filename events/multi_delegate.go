package events

import "github.com/hx/midground"

// MultiDelegate is a midground.Delegate implementation that re-sends notifications to multiple delegates.
type MultiDelegate []midground.Delegate

func (m MultiDelegate) JobScheduled(process *midground.Process) {
	for _, d := range m {
		if d != nil {
			d.JobScheduled(process)
		}
	}
}

func (m MultiDelegate) JobBlocked(process *midground.Process, blockers []*midground.Process) {
	for _, d := range m {
		if d != nil {
			d.JobBlocked(process, blockers)
		}
	}
}

func (m MultiDelegate) JobStarting(process *midground.Process) {
	for _, d := range m {
		if d != nil {
			d.JobStarting(process)
		}
	}
}

func (m MultiDelegate) JobProgressed(process *midground.Process, payload any) {
	for _, d := range m {
		if d != nil {
			d.JobProgressed(process, payload)
		}
	}
}

func (m MultiDelegate) JobEnded(process *midground.Process, err error) {
	for _, d := range m {
		if d != nil {
			d.JobEnded(process, err)
		}
	}
}
