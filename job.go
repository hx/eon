package midground

type Job struct {
	Tag        any
	Runner     func(ctx *Context) error
	Supersedes func(enqueued []*Process) (superseded []*Process)
	Readiness  func(running []*Process) Readiness
}
