package eon

// ErrSuperseded is passed to Delegate.JobEnded when a Process is by Replacement's Job.Supersedes function before it can
// start.
type ErrSuperseded struct {
	Replacement *Process
}

func (e ErrSuperseded) Error() string { return "superseded" }

// ErrDiscarded is passed to Delegate.JobEnded when a Process is discarded by its Job's Readiness function.
type ErrDiscarded struct{}

func (e ErrDiscarded) Error() string { return "discarded" }

// ErrContextExpired is passed to Delegate.JobEnded when a Process's context expires before it starts.
type ErrContextExpired struct {
	error
}

func (e ErrContextExpired) Unwrap() error { return e.error }

// ErrSchedulerContextExpired is passed to Delegate.JobEnded for every unstarted Process of a Scheduler whose context
// has expired.
type ErrSchedulerContextExpired struct {
	ErrContextExpired
}

func (e ErrSchedulerContextExpired) Unwrap() error { return e.error }
