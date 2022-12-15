package midground

import (
	"fmt"
	"testing"
)

type customDelegate struct{}

func (d customDelegate) JobScheduled(process *Process)                    {}
func (d customDelegate) JobBlocked(process *Process, blockers []*Process) {}
func (d customDelegate) JobStarting(process *Process)                     { _ = fmt.Sprintf("%#v", struct{}{}) }
func (d customDelegate) JobProgressed(process *Process, payload any)      {}
func (d customDelegate) JobEnded(process *Process, err error)             {}

func BenchmarkDelegation(b *testing.B) {
	var (
		process   = &Process{}
		withDepth = func(depth int) (n *Dispatcher) {
			n = &Dispatcher{Delegate: customDelegate{}}
			for i := 1; i < depth; i++ {
				n = &Dispatcher{Delegate: customDelegate{}, process: &Process{dispatcher: n}}
			}
			return
		}
	)
	for _, depth := range []struct {
		depth      string
		dispatcher *Dispatcher
	}{
		{"single", withDepth(1)},
		{"double", withDepth(2)},
		{"triple", withDepth(3)},
		{"ten", withDepth(10)},
	} {
		e := depth.dispatcher
		b.Run(depth.depth+"_depth", func(b *testing.B) {
			b.Run("notify", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					e.notify(func(d Delegate) { d.JobStarting(process) })
				}
			})
			// b.Run("chain", func(b *testing.B) {
			// 	for i := 0; i < b.N; i++ {
			// 		for _, d := range e.delegateChain() {
			// 			d.JobStarting(process)
			// 		}
			// 	}
			// })
		})
	}
}
