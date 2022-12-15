package midground

type ErrSuperseded struct {
	Replacement *Process
}

func (e ErrSuperseded) Error() string { return "superseded" }

type ErrDiscarded struct{}

func (e ErrDiscarded) Error() string { return "discarded" }
