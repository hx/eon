package midground

import "context"

type Context struct {
	*Dispatcher
	ctx context.Context
}

func (c *Context) Ctx() context.Context {
	if c.ctx == nil {
		return context.Background()
	}
	return c.ctx
}

func (c *Context) Progress(payload any) {
	// TODO: interrupt scheduler?
	c.notify(func(d Delegate) { d.JobProgressed(c.process, payload) })
}
