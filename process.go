package midground

import (
	"context"
	"time"
)

// A Process is a Job that has been scheduled.
type Process struct {
	Job        *Job
	schedule   *schedule
	runAt      time.Time
	batch      *batch
	batchIndex int
	state      int64
	ctx        context.Context
	dispatcher *Dispatcher
}

const (
	stateScheduled int64 = iota
	stateBlocked
	stateRunning
	stateEnded
)

// Parent is the Process whose Context was used to schedule this Process. If the Process was scheduled directly on a
// Scheduler, Parent will be nil.
func (p *Process) Parent() *Process { return p.dispatcher.process }

// IsScheduled is true when the Process is not yet due to run, or has not yet had its Readiness evaluated.
func (p *Process) IsScheduled() bool { return p.state == stateScheduled }

// IsBlocked is true when the Process is waiting for another Process to end.
func (p *Process) IsBlocked() bool { return p.state == stateBlocked }

// IsRunning is true when the Process is running.
func (p *Process) IsRunning() bool { return p.state == stateRunning }

// IsEnded is true when the Process is finished, has been superseded, or has an expired context.
func (p *Process) IsEnded() bool { return p.state == stateEnded }

func (p *Process) readiness(running []*Process) (readiness Readiness) {
	if p.Job.Readiness != nil {
		readiness = p.Job.Readiness(running)
	}
	if readiness.isDiscarded {
		return
	}
	if blocker := p.batch.findBlocker(p); blocker != nil {
		for _, other := range readiness.blockers {
			if other == blocker {
				return
			}
		}
		readiness.blockers = append(readiness.blockers, blocker)
	}
	return
}

func (p *Process) terminate(err error) {
	p.state = stateEnded
	p.dispatcher.notify(func(d Delegate) { d.JobEnded(p, err) })
	p.batch.done()
}
