package events

import "github.com/hx/eon"

// Delegate is a midground.Delegate implementation that allows each event handler to be optionally specified as an
// anonymous function.
type Delegate struct {
	Scheduled  func(process *eon.Process)
	Blocked    func(process *eon.Process, blockers []*eon.Process)
	Starting   func(process *eon.Process)
	Progressed func(process *eon.Process, payload any)
	Ended      func(process *eon.Process, err error)
}

func (d *Delegate) JobScheduled(process *eon.Process) {
	if d != nil && d.Scheduled != nil {
		d.Scheduled(process)
	}
}

func (d *Delegate) JobBlocked(process *eon.Process, blockers []*eon.Process) {
	if d != nil && d.Blocked != nil {
		d.Blocked(process, blockers)
	}
}

func (d *Delegate) JobStarting(process *eon.Process) {
	if d != nil && d.Starting != nil {
		d.Starting(process)
	}
}

func (d *Delegate) JobProgressed(process *eon.Process, payload any) {
	if d != nil && d.Progressed != nil {
		d.Progressed(process, payload)
	}
}

func (d *Delegate) JobEnded(process *eon.Process, err error) {
	if d != nil && d.Ended != nil {
		d.Ended(process, err)
	}
}
