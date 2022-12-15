package midground

import (
	"context"
	"time"
)

// batch represents a series of jobs that were scheduled at the same time. It is used to ensure they run in the correct
// order, and that the concurrency limit is honoured.
type batch struct {
	processes   []*Process
	concurrency int
	doneCount   int
	doneChan    chan struct{}
}

// newBatch creates a new batch, including new Process values, from the given slice of Job values.
func newBatch(ctx context.Context, owner *Dispatcher, concurrency int, schedule *schedule, now time.Time, jobs []*Job) (b *batch) {
	runAt := now
	if schedule != nil {
		runAt = runAt.Add(schedule.delay)
	}
	if concurrency == 0 {
		concurrency = len(jobs)
	}
	b = &batch{
		concurrency: concurrency,
		doneChan:    make(chan struct{}),
	}
	b.processes = make([]*Process, len(jobs))
	for i, job := range jobs {
		process := &Process{
			Job:        job,
			schedule:   schedule,
			batch:      b,
			batchIndex: i,
			runAt:      runAt,
			state:      stateScheduled,
			ctx:        ctx,
			dispatcher: owner,
		}
		b.processes[i] = process
	}
	return
}

// done decreases doneCount, and closes doneChan when it reaches the number of processes in the batch.
func (s *batch) done() {
	s.doneCount++
	if s.doneCount == len(s.processes) {
		close(s.doneChan)
	}
}

// findBlocker returns the process from this sequence that is most directly blocking the given process. If it is not
// blocked, nil is returned.
func (s *batch) findBlocker(process *Process) *Process {
	// TODO: redo this using batchIndex
	if len(s.processes) == 1 {
		return nil
	}
	var runningCount int
	for i, otherProcess := range s.processes {
		if otherProcess == process {
			if i == 0 {
				return nil
			}
			previousProcess := s.processes[i-1]
			if previousProcess.IsScheduled() || previousProcess.IsBlocked() || runningCount == s.concurrency {
				return previousProcess
			}
			return nil
		} else if otherProcess.IsRunning() {
			runningCount++
		}
	}
	panic("job not in batch")
}
