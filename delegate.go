package eon

// Delegate implementations can be assigned to any Dispatcher (i.e. a Scheduler or a Context), to receive notifications
// about changes in Process state.
type Delegate interface {
	// JobScheduled is invoked when Job is scheduled as a new Process, and is waiting to start.
	JobScheduled(process *Process)

	// JobBlocked is invoked when a Job's Process is due to run, but is waiting for another Process(es) to finish.
	JobBlocked(process *Process, blockers []*Process)

	// JobStarting is invoked just before a Job's Process is started.
	JobStarting(process *Process)

	// JobProgressed is invoked whenever Context.Progress is called from a running Process.
	JobProgressed(process *Process, payload any)

	// JobEnded is invoked whenever a Job's Process ends, whether by completing, being superseded, or having its
	// context cancelled. Every invocation of JobScheduled will eventually be followed by JobEnded.
	JobEnded(process *Process, err error)
}
