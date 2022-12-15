package midground

import (
	"context"
	"time"
)

type Process struct {
	Job        *Job
	schedule   *Schedule
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

func (p *Process) Parent() *Process { return p.dispatcher.process }

func (p *Process) IsScheduled() bool { return p.state == stateScheduled }
func (p *Process) IsBlocked() bool   { return p.state == stateBlocked }
func (p *Process) IsRunning() bool   { return p.state == stateRunning }
func (p *Process) IsEnded() bool     { return p.state == stateEnded }

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
