package midground

import (
	"context"
	"github.com/hx/midground/util"
	"time"
)

type action struct {
	work func()
	done chan struct{}
}

type Scheduler struct {
	*Dispatcher
	ctx       context.Context
	scheduled *util.Heap[*Process]
	blocked   util.Set[*Process]
	running   []*Job
	actions   chan chan action
	clock     clock
}

func NewScheduler() *Scheduler { return NewSchedulerContext(context.Background()) }

func NewSchedulerContext(ctx context.Context) (scheduler *Scheduler) {
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

func (s *Scheduler) schedule(ctx context.Context, owner *Dispatcher, concurrency int, schedule *Schedule, wait bool, jobs []*Job) {
	if s.ctx.Err() != nil {
		panic("root context done")
	}
	if len(jobs) == 0 {
		return
	}
	b := s.makeBatch(ctx, owner, concurrency, schedule, jobs)
	<-s.interrupt(func() {
		for _, process := range b.processes {
			s.scheduleProcess(process)
		}
	})
	s.subscribeToContext(ctx, b)
	if wait {
		<-b.doneChan
	}
}

func (s *Scheduler) reschedule(process *Process) {
	delay := process.schedule.RepeatAfter
	schedule := &Schedule{
		Delay:       delay,
		RepeatAfter: delay,
	}
	b := s.makeBatch(process.ctx, process.dispatcher, 1, schedule, []*Job{process.Job})
	s.scheduleProcess(b.processes[0])
	s.subscribeToContext(process.ctx, b)
}

func (s *Scheduler) scheduleProcess(process *Process) {
	s.scheduled.Push(process)
	process.dispatcher.notify(func(d Delegate) { d.JobScheduled(process) })
}

func (s *Scheduler) subscribeToContext(ctx context.Context, b *batch) {
	// TODO: make this configurable, since it's potentially so expensive/messy
	if ctx != nil {
		go func() {
			select {
			case <-b.doneChan:
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

func (s *Scheduler) makeBatch(ctx context.Context, owner *Dispatcher, concurrency int, schedule *Schedule, jobs []*Job) (b *batch) {
	var runAt = s.clock.Now()
	if schedule != nil {
		runAt = runAt.Add(schedule.Delay)
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

func (s *Scheduler) start(process *Process) {
	if process.Job.Supersedes != nil {
		s.removeSuperseded(process.Job.Supersedes(s.queuedProcesses()), process)
	}
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
			if process.schedule != nil && process.ctx.Err() == nil && process.schedule.RepeatAfter > 0 {
				s.reschedule(process)
			}
		})
	}()
}

func (s *Scheduler) queuedProcesses() []*Process {
	return append(s.blocked.Slice(), s.scheduled.Slice()...)
}

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

func (s *Scheduler) interrupt(fn func()) (done chan struct{}) {
	done = make(chan struct{})
	<-s.actions <- action{fn, done}
	return
}

func (s *Scheduler) findRunnableProcess(now time.Time) (process *Process) {
	// search blocked
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

	// check waiting
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

func (s *Scheduler) cancel(process *Process, err error) {
	if process.IsBlocked() {
		s.blocked.Remove(process)
	} else if process.IsScheduled() {
		s.scheduled.Remove(process)
	}
	process.terminate(err)
}

func (s *Scheduler) nextScheduledProcess(now time.Time) (process *Process, runNow bool) {
	process = s.scheduled.Peek()
	if process == nil {
		return
	}
	runNow = now.Before(process.runAt)
	if runNow {
		s.scheduled.Pop()
	}
	return
}
