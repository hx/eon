package midground

type Delegate interface {
	JobScheduled(process *Process)
	JobBlocked(process *Process, blockers []*Process)
	JobStarting(process *Process)
	JobProgressed(process *Process, payload any)
	JobEnded(process *Process, err error)
}
