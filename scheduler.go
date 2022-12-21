package eon

import (
	"context"
	"github.com/hx/eon/util"
	"time"
)

// action is a unit of work a Scheduler performs on itself when it interrupts its run loop.
type action struct {
	// work is the work to be performed
	work func()

	// done is returned by Scheduler.interrupt, and can be used to wait for the action to complete, if the caller needs
	// to.
	done chan struct{}
}

// A Scheduler is the top level Dispatcher for a group of processes.
//
// A zero value is not usable. Use NewScheduler instead.
type Scheduler struct {
	*Dispatcher
	ctx       context.Context
	scheduled *util.Heap[*Process]
	blocked   util.Set[*Process]
	running   []*Job // TODO: consider a util.BlockingHeap instead so that loop can wait for all jobs to finish
	actions   chan chan action
	clock     clock
}

// NewScheduler returns a ready-to-use Scheduler. It starts a scheduling goroutine that will run until the given
// context expires.
func NewScheduler(ctx context.Context) (scheduler *Scheduler) {
	scheduler = &Scheduler{
		clock:     systemClock{},
		actions:   make(chan chan action),
		scheduled: util.NewHeap(func(a, b *Process) bool { return a.runAt.Before(b.runAt) }),
		blocked:   util.NewSet[*Process](),
		ctx:       ctx,
	}
	scheduler.Dispatcher = &Dispatcher{
		scheduler: scheduler,
	}
	go scheduler.loop()
	return
}

// loop is the main run loop for its receiver. It runs in a goroutine started by the constructor, NewScheduler.
func (s *Scheduler) loop() {
	var now time.Time
	for s.ctx.Err() == nil {
		now = s.clock.Now()
		if job := s.findRunnableProcess(now); job != nil {
			s.start(job)
		} else {
			s.idle(now)
		}
	}
	for _, job := range s.queuedProcesses() {
		s.cancel(job, s.ctx.Err())
	}
	// TODO: wait for running processes to finish? add indirection to consumption of <-s.ctx.Done() in idle()
}

// interrupt runs the given function on the receiver's run loop. It will deadlock if called recursively. To wait for the
// given function to complete, read from the returned channel.
func (s *Scheduler) interrupt(fn func()) (done chan struct{}) {
	done = make(chan struct{})
	<-s.actions <- action{fn, done}
	return
}

// schedule is used by Dispatcher to scheduler new jobs.
func (s *Scheduler) schedule(ctx context.Context, owner *Dispatcher, concurrency int, schedule *schedule, jobs []*Job) <-chan struct{} {
	if s.ctx.Err() != nil {
		panic("root context expired")
	}
	b := newBatch(ctx, owner, concurrency, schedule, s.clock.Now(), jobs)
	if len(jobs) > 0 {
		<-s.interrupt(func() {
			for _, process := range b.processes {
				if process.Job.Supersedes != nil {
					s.removeSuperseded(process.Job.Supersedes(s.queuedProcesses()), process)
				}
				s.scheduleProcess(process)
			}
		})
		s.subscribeToContext(ctx, b)
	}
	return b.doneChan
}

// reschedule is invoked after a process has ended if it has a schedule with a positive repeatAfter value. It must be
// run from an interrupt function.
func (s *Scheduler) reschedule(process *Process) {
	delay := process.schedule.repeatAfter
	schedule := &schedule{
		delay:       delay,
		repeatAfter: delay,
	}
	b := newBatch(process.ctx, process.dispatcher, 1, schedule, s.clock.Now(), []*Job{process.Job})
	s.scheduleProcess(b.processes[0])
	s.subscribeToContext(process.ctx, b)
}

// scheduleProcess adds the given process to the receiver's schedule. It must be run from an interrupt function.
func (s *Scheduler) scheduleProcess(process *Process) {
	s.scheduled.Push(process)
	process.dispatcher.notify(func(d Delegate) { d.JobScheduled(process) })
}

// subscribeToContext starts a goroutine that cancels the given batch if the given context expires.
func (s *Scheduler) subscribeToContext(ctx context.Context, b *batch) {
	// TODO: make this configurable, since it's potentially so expensive/messy
	if ctx != nil {
		go func() {
			select {
			case <-b.doneChan: // TODO: It's not necessary to wait for all jobs to finish. We only need to wait for them to start.
			case <-ctx.Done():
				s.interrupt(func() {
					for _, job := range b.processes {
						if job.IsScheduled() || job.IsBlocked() {
							s.cancel(job, ctx.Err())
						}
					}
				})
			}
		}()
	}
}

// idle is used by the run loop to wait for something to happen; either a scheduled job falls due, the loop is
// interrupted by the interrupt method, or the receiver's context expires.
func (s *Scheduler) idle(now time.Time) {
	var (
		job       = s.scheduled.Peek()
		interrupt = make(chan action)
		timeout   time.Duration
		act       action
	)
	if job != nil {
		timeout = job.runAt.Sub(now)
	}
	if timeout > 0 {
		select {
		case <-s.clock.After(timeout):
			return
		case <-s.ctx.Done():
			return
		case s.actions <- interrupt:
			act = <-interrupt
		}
	} else {
		select {
		case <-s.ctx.Done():
			return
		case s.actions <- interrupt:
			act = <-interrupt
		}
	}
	act.work()
	close(act.done)
}

// start runs the given Process, by first cancelling superseded processes, then running it in a new goroutine, which
// also takes care of termination and rescheduling of repeated jobs.
func (s *Scheduler) start(process *Process) {
	s.running = append(s.running, process.Job)
	process.state = stateRunning
	process.dispatcher.notify(func(d Delegate) { d.JobStarting(process) })
	ctx := &Context{
		Dispatcher: &Dispatcher{
			scheduler: s,
			process:   process,
		},
	}
	go func() {
		err := process.Job.Runner(ctx)
		s.interrupt(func() {
			util.Remove(&s.running, process.Job)
			process.terminate(err)
			if process.schedule != nil && process.schedule.repeatAfter > 0 && process.ctx.Err() == nil {
				s.reschedule(process)
			}
		})
	}()
}

// queuedProcesses returns a slice of all queued processes, including scheduled and blocked processes.
func (s *Scheduler) queuedProcesses() []*Process {
	return append(s.blocked.Slice(), s.scheduled.Slice()...)
}

// removeSuperseded is used by Scheduler.schedule to remove superseded processes from the schedule.
func (s *Scheduler) removeSuperseded(supersededProcesses []*Process, replacement *Process) {
	if len(supersededProcesses) == 0 {
		return
	}
	supersededJobSet := util.NewSet(supersededProcesses...)
	for _, job := range append(s.blocked.Slice(), s.scheduled.Slice()...) {
		if supersededJobSet.Contains(job) {
			s.cancel(job, ErrSuperseded{replacement})
		}
	}
}

// findRunnableProcess is used by the run loop to look for processes to schedule.
func (s *Scheduler) findRunnableProcess(now time.Time) (process *Process) {
	// search blocked processes first
	for process = range s.blocked {
		readiness := process.readiness(s.queuedProcesses())
		if readiness.isDiscarded {
			// discard
			s.blocked.Remove(process)
			process.terminate(ErrDiscarded{})
		} else if len(readiness.blockers) == 0 {
			// ready
			s.blocked.Remove(process)
			return
		} else {
			// still blocked, do nothing
			process.dispatcher.notify(func(d Delegate) { d.JobBlocked(process, readiness.blockers) })

		}
	}

	// if nothing is blocked, look for scheduled process that are due or past due
	for {
		process = s.scheduled.Peek()
		if process == nil {
			return
		}
		if process.runAt.After(now) {
			return nil
		}
		s.scheduled.Pop()
		readiness := process.readiness(s.queuedProcesses())
		if readiness.isDiscarded {
			// discard
			process.terminate(ErrDiscarded{})
			continue
		} else if len(readiness.blockers) == 0 {
			// ready
			return
		} else {
			// blocked
			process.state = stateBlocked
			s.blocked.Add(process)
			process.dispatcher.notify(func(d Delegate) { d.JobBlocked(process, readiness.blockers) })
		}
	}
}

// cancel removes and terminates a process that hasn't started yet.
func (s *Scheduler) cancel(process *Process, err error) {
	if process.IsBlocked() {
		s.blocked.Remove(process)
	} else if process.IsScheduled() {
		s.scheduled.Remove(process)
	}
	process.terminate(err)
}
