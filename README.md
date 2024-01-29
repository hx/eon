# Eon

**Task scheduling for Golang**

```plain
go get github.com/hx/eon
```

[![GoDoc](https://godoc.org/github.com/hx/eon?status.svg)](https://godoc.org/github.com/hx/eon)

## Overview

Eon is an in-process task scheduler. Its key advantages include:

- Delay and repeat jobs.
- Debounce and dedupe multiple requests for the same job.
- Supersede jobs with other jobs (e.g. "reload all" replaces "reload one").
- Finely tunable concurrency limits with FIFO queuing.
- Events for scheduling, starting, progressing, and finishing jobs, providing flexible UI integration.
- Jobs can enqueue child jobs, complete with event bubbling. 
- Context-based cancellation at any level, including the root scheduler.

## Usage

Typical usage of Eon involves a single `Scheduler` spanning the lifetime of a process.

```go
package main

import (
    "context"
    "fmt"
    "github.com/hx/eon"
    "os/signal"
    "syscall"
)

func main() {
    ctx, _ := signal.NotifyContext(
        context.Background(),
        syscall.SIGINT,
        syscall.SIGTERM,
    )

    scheduler := eon.NewScheduler(ctx)

    scheduler.Run(ctx, &eon.Job{
        Runner: func(ctx *eon.Context) error {
            _, err := fmt.Println("Hello, world!")
            return err
        },
    })

    <-ctx.Done() // Wait for an INT or TERM signal
}
```

### Timing

Jobs can be scheduled with delays.

To delay a job by 10 seconds:

```go
scheduler.Schedule(ctx, 10*time.Second, 0, &eon.Job{ … })
```

Jobs can also be made to self-reschedule on completion, by specifying a positive `repeatAfter` value.

To repeat the above job with one-minute pauses:

```go
scheduler.Schedule(ctx, 10*time.Second, time.minute, &eon.Job{ … })
```

This is not the same as a one-minute _interval_. It is the amount of time between a process finishing and the next process starting. The total interval is the `repeatAfter` value plus the duration of the process itself. 

Jobs that return errors are not rescheduled.

### Parallelism

The `Run` function in the example above accepts one or more `Job` values, and runs them one after the other. It is equivalent to calling the `Parallel` function with a concurrency value of `1`.

You can also specify maximum concurrency:

```go
jobs := make([]*eon.Job, len(users))
for i, user := range users {
    jobs[i] = reapExpiredAuthTokens(user)
}

// Process a maximum of 3 users at once
scheduler.Parallel(ctx, 3, jobs...)
```

Specify concurrency as `0` to set no limit.

### Child jobs

Processes (running jobs) can schedule children using their `Context` arguments.

```go
scheduler.Run(
    ctx,
    &eon.Job{
        Tag: "parent job",
        Runner: func(ctx *eon.Context) error {
            ctx.Run(
                ctx.Ctx(),
                &eon.Job{
                    Tag: "child job",
                    Runner: func(ctx *eon.Context) error {
                        …
                    },
                },
            )

            return nil
        },
    },
)
```

### Monitoring

To monitor the activity of a `Dispatcher`, assign a `Delegate`. Package `events` contains several implementations, or you can use your own.

```go
scheduler.Delegate = &events.Delegate{
    Scheduled: func(process *eon.Process) {
        fmt.Printf("Process %v scheduled.\n", process.Job.Tag)
    },
    Ended: func(process *eon.Process, err error) {
        fmt.Printf("Process %v ended with %v.\n", process.Job.Tag, err)
    },
}
```

Other events include `Blocked`, `Starting`, and `Progressed`. See API docs for more information.

Although processes run on their own goroutines, schedulers run on a single goroutine. This includes calls to `Delegate` methods, which ensures they are invoked in the expected order. If you want to run expensive operations in delegates, consider using extra goroutines so as not to block the scheduler.

Delegates can also be assigned to the `Context` of any running process. Delegates are invoked in ascending order, with the root (scheduler) delegate being invoked last.

### Exclusivity

Specifying a `Readiness` function allows jobs to block, particularly based on running jobs (processes).

```go
job := &eon.Job{
    Tag: SomeType{},
    Runner: func(ctx *eon.Context) error {
        …
    },
    Readiness: func(running []*eon.Process) eon.Readiness {
        for _, process := range running {
            if _, ok := process.Job.Tag.(SomeType); ok {
                return eon.Blocked(process)
            }
        }
        return eon.Ready()
    },
}
```

`Readiness` functions may return one of:

* `Ready()`, indicating that the job is ready to run;
* `Discard()`, indicating the job should be discarded (e.g. as a duplicate); or
* `Blocked(processes...)`, indicating the job conflicts with one or more running processes. Given `processes` do not affect behaviour, but are included in `Blocked` events.

### Supersession

Before a job starts, it has an opportunity to supersede other scheduled jobs.

```go
scheduler.Run(ctx, &eon.Job{
    Tag:    "delete all access tokens",
    Runner: func(ctx *eon.Context) error { … },
    Supersedes: func(enqueued []*eon.Process) (superseded []*eon.Process) {
        for _, process := range enqueued {
            if process.Job.Tag == "delete all access tokens" || 
                process.Job.Tag == "delete expired access tokens" {
                superseded = append(superseded, process)
            }
        }
        return
    },
})
```

## Common behaviours

### Cancel on error

It is often beneficial to halt a sequence of jobs after a job fails.

Eon does not provide an opinionated means to achieve this, since there are enumerable nuances to error propagation, application-specific job wrappers, and approaches to concurrency.

A "run jobs until one fails" function in your application is a reasonable component of a robust solution, and context cancellation is generally the right tool for the job, as demonstrated in this example:

```go
type BatchHalted struct {
    FailedJobIndex int
    Jobs           []*eon.Job
    Err            error
}

func (b BatchHalted) Error() string {
    return fmt.Sprintf("batch halted at job %d of %d: %s", b.FailedJobIndex+1, len(b.Jobs), b.Err)
}

func (b BatchHalted) Unwrap() error { return b.Err }

func RunUntilError(ctx context.Context, dispatcher *eon.Dispatcher, jobs ...*eon.Job) {
    var (
        c, cancel        = context.WithCancelCause(ctx)
        originalDelegate = dispatcher.Delegate
        wg               sync.WaitGroup
        completed int
    )
    wg.Add(len(jobs))
    dispatcher.Delegate = events.MultiDelegate{originalDelegate, &events.Delegate{
        Ended: func(process *eon.Process, err error) {
            cancel(BatchHalted{completed, jobs, err}) // no-op after first call
            wg.Done()
            completed++
        },
    }}
    dispatcher.Run(c, jobs...)
    
    // Here, blocking on the wait group allows synchronous execution in the caller.
    // It also restores the delegate at the correct time. This function could be
    // made asynchronous, in which case, using an events.Handler instead of
    // events.MultiDelegate would be a more suitable concurrency-safe alternative.
    wg.Wait()
    dispatcher.Delegate = originalDelegate
}
```

### TUI reporting

TODO

