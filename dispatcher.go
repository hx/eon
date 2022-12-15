package midground

import (
	"context"
)

type scheduler interface {
	schedule(ctx context.Context, owner *Dispatcher, concurrency int, schedule *Schedule, wait bool, jobs []*Job)
}

type Dispatcher struct {
	Delegate Delegate

	scheduler scheduler

	// The Process for which the Dispatcher was created, or nil if it is a root dispatcher (belonging to a Scheduler).
	process *Process // nil for root
}

func (d *Dispatcher) Run(jobs ...*Job) {
	d.RunContext(nil, jobs...)
}

func (d *Dispatcher) RunContext(ctx context.Context, jobs ...*Job) {
	d.scheduler.schedule(ctx, d, 1, nil, true, jobs)
}

func (d *Dispatcher) Parallel(concurrency int, jobs ...*Job) {
	d.ParallelContext(nil, concurrency, jobs...)
}

func (d *Dispatcher) ParallelContext(ctx context.Context, concurrency int, jobs ...*Job) {
	d.scheduler.schedule(ctx, d, concurrency, nil, true, jobs)
}

func (d *Dispatcher) Schedule(schedule *Schedule, jobs ...*Job) {
	d.ScheduleContext(nil, schedule, jobs...)
}

func (d *Dispatcher) ScheduleContext(ctx context.Context, schedule *Schedule, jobs ...*Job) {
	for _, job := range jobs {
		d.scheduler.schedule(ctx, d, 1, schedule, false, []*Job{job})
	}
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
