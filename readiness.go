package midground

// Readiness is the type returned by Job.Readiness that indicates whether a job is ready to run. Its zero value is
// equivalent to Ready(). Other values can be declared with Blocked(blockers...) and Discard().
type Readiness struct {
	isDiscarded bool
	blockers    []*Process
}

// Blocked indicates that a job is blocked by another job or jobs, and should wait for them to finish.
func Blocked(blockers ...*Process) Readiness { return Readiness{blockers: blockers} }

// Discard indicates that a job should not run, and should be discarded. If it is a repeating job, it will not be
// rescheduled.
func Discard() Readiness { return Readiness{isDiscarded: true} }

// Ready indicates that a job is ready to run.
func Ready() Readiness { return Readiness{} }
