package eon

import (
	"context"
	"time"
)

// scheduler is used to avoid a two-way dependency between Scheduler and Dispatcher. It's probably unnecessary.
type scheduler interface {
	schedule(ctx context.Context, owner *Dispatcher, concurrency int, schedule *schedule, jobs []*Job) <-chan struct{}
}

// A Dispatcher can be used to schedule new jobs, and track their progress through its Delegate.
//
// Dispatcher is used by the Scheduler and Context types. It is not intended to be used on its own. It can, however, be
// pulled out and passed to more complex scheduling functions that are intended to work on either Scheduler or Context
// values.
type Dispatcher struct {
	// Delegate receives notifications whenever a Process started by the Dispatcher, or one of its descendents, is
	// scheduled, blocked, starting, progressed, or ended.
	//
	// Care should be taken to assign Delegate when no child processes are running, to avoid races. If it is necessary
	// to add and/or remove event handlers while processes are running, consider using events.Handler as your Delegate,
	// which can be safely modified from multiple goroutines.
	Delegate Delegate

	scheduler scheduler

	// The Process for which the Dispatcher was created, or nil if it is a root dispatcher (belonging to a Scheduler).
	process *Process // nil for root
}

// Run schedules the given jobs to run one after the other, and blocks until they have all ended.
//
// A failing job will not prevent subsequent jobs from running. To halt a batch of jobs on error, use Run with
// a context that can be cancelled by Delegate.JobEnded when it receives an error.
//
// Internally, handling each context requires an extra goroutine to live from when the given batch is scheduled until
// the last job starts. If you don't need to be able to stop your jobs, leave ctx as nil to save the cost of a
// goroutine.
func (d *Dispatcher) Run(ctx context.Context, jobs ...*Job) {
	d.Parallel(ctx, 1, jobs...)
}

// Parallel is identical to Run, but allows maximum concurrency to be specified for the batch.
func (d *Dispatcher) Parallel(ctx context.Context, concurrency int, jobs ...*Job) {
	<-d.scheduler.schedule(ctx, d, concurrency, nil, jobs)
}

// Schedule schedules a job to be run in the background, with an optional delay. A positive repeatAfter value will
// reschedule the job upon its completion with a new delay.
//
// Internally, handling each context requires an extra goroutine to live from when the given job is scheduled until
// it starts. If you don't need to be able to stop your job, leave ctx as nil to save the cost of a goroutine.
func (d *Dispatcher) Schedule(ctx context.Context, delay, repeatAfter time.Duration, job *Job) {
	d.scheduler.schedule(ctx, d, 1, &schedule{delay, repeatAfter}, []*Job{job})
}

/**

A non-recursive version of this function benchmarked slightly slower for dispatchers with no parents, and around the
same speed for dispatchers with ancestors. I also tried building a delegate chain and having the caller loop over it, to
avoid going full FP, but that was 25% to 35% slower, chiefly because the delegate chain has to be rebuilt each time to
allow for changes to the value of Dispatcher.Delegate.

-benchtime=2s -count=2

goos: darwin
goarch: amd64
pkg: github.com/hx/midground
cpu: Intel(R) Core(TM) i7-9750H CPU @ 2.60GHz
BenchmarkDelegation/single_depth/notify-12  	23344108	        98.32 ns/op
BenchmarkDelegation/single_depth/notify-12  	24168766	        97.15 ns/op
BenchmarkDelegation/single_depth/chain-12   	17657547	       135.6 ns/op
BenchmarkDelegation/single_depth/chain-12   	17581126	       135.5 ns/op
BenchmarkDelegation/double_depth/notify-12  	11903797	       205.1 ns/op
BenchmarkDelegation/double_depth/notify-12  	11654077	       204.9 ns/op
BenchmarkDelegation/double_depth/chain-12   	 7673158	       312.9 ns/op
BenchmarkDelegation/double_depth/chain-12   	 7640418	       311.9 ns/op
BenchmarkDelegation/triple_depth/notify-12  	 8103494	       294.4 ns/op
BenchmarkDelegation/triple_depth/notify-12  	 8114937	       294.7 ns/op
BenchmarkDelegation/triple_depth/chain-12   	 4966002	       478.2 ns/op
BenchmarkDelegation/triple_depth/chain-12   	 5016427	       484.1 ns/op
BenchmarkDelegation/ten_depth/notify-12     	 2512461	       952.1 ns/op
BenchmarkDelegation/ten_depth/notify-12     	 2517297	       956.4 ns/op
BenchmarkDelegation/ten_depth/chain-12      	 1796907	      1335 ns/op
BenchmarkDelegation/ten_depth/chain-12      	 1795909	      1337 ns/op

*/

func (d *Dispatcher) notify(fn func(d Delegate)) {
	if d.Delegate != nil {
		fn(d.Delegate)
	}
	if d.process != nil {
		d.process.dispatcher.notify(fn)
	}
}
