package midground

// ErrSuperseded is passed to Delegate.JobEnded when a Process is by Replacement's Job.Supersedes function before it can
// start.
type ErrSuperseded struct {
	Replacement *Process
}

func (e ErrSuperseded) Error() string { return "superseded" }

// ErrDiscarded is passed to Delegate.JobEnded when a Process is discarded by its Job's Readiness function.
type ErrDiscarded struct{}

func (e ErrDiscarded) Error() string { return "discarded" }
