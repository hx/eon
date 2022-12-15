package midground

// A Job is a single unit of work. A Job becomes a Process when it is scheduled to run with a Scheduler or Context.
//
// Jobs should be considered immutable. You can safely schedule a Job more than once, on multiple
type Job struct {
	// Runner is the work to be performed by the Job.
	Runner func(ctx *Context) error

	// Tag can be used by the Job creator to attach whatever data they need to a Job. It can be used in Delegate methods
	// to report activity, and in Supersedes and Readiness functions to identify jobs in other processes. It has no
	// meaning in this package.
	Tag any

	// Supersedes, if not nil, is invoked when a Process is about to start, with a slice of other processes that have
	// not yet started. It should return a (possibly empty) subset of that slice containing processes that are to be
	// superseded by the receiver's process.
	Supersedes func(enqueued []*Process) (superseded []*Process)

	// Readiness, if not nil, is invoked when a Process is about to start, and should return one of Ready(), Discard(),
	// or Blocked(blockers...), where blockers is a subset of the passed slice of running processes. If nil, the Job is
	// assumed to always be ready.
	Readiness func(running []*Process) Readiness
}
