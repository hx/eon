package eon

import "context"

// A Context is created for every running Process.
type Context struct {
	*Dispatcher
	ctx context.Context
}

// Ctx is the context.Context with which the process's Job was scheduled.
func (c *Context) Ctx() context.Context {
	if c.ctx == nil {
		return context.Background()
	}
	return c.ctx
}

// Progress reports the process's process. The given payload is passed to the Delegate.JobProgressed function of
// delegates in the receiver and its ancestors.
func (c *Context) Progress(payload any) {
	// TODO: interrupt scheduler?
	c.notify(func(d Delegate) { d.JobProgressed(c.process, payload) })
}
