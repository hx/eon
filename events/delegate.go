package events

import "github.com/hx/midground"

// Delegate is a midground.Delegate implementation that allows each event handler to be optionally specified as an
// anonymous function.
type Delegate struct {
	Scheduled  func(process *midground.Process)
	Blocked    func(process *midground.Process, blockers []*midground.Process)
	Starting   func(process *midground.Process)
	Progressed func(process *midground.Process, payload any)
	Ended      func(process *midground.Process, err error)
}

func (d *Delegate) JobScheduled(process *midground.Process) {
	if d != nil && d.Scheduled != nil {
		d.Scheduled(process)
	}
}

func (d *Delegate) JobBlocked(process *midground.Process, blockers []*midground.Process) {
	if d != nil && d.Blocked != nil {
		d.Blocked(process, blockers)
	}
}

func (d *Delegate) JobStarting(process *midground.Process) {
	if d != nil && d.Starting != nil {
		d.Starting(process)
	}
}

func (d *Delegate) JobProgressed(process *midground.Process, payload any) {
	if d != nil && d.Progressed != nil {
		d.Progressed(process, payload)
	}
}

func (d *Delegate) JobEnded(process *midground.Process, err error) {
	if d != nil && d.Ended != nil {
		d.Ended(process, err)
	}
}
