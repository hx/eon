package events_test

import (
	"fmt"
	"github.com/hx/eon"
	"github.com/hx/eon/events"
	"testing"
)

type customDelegate struct{}

func (d customDelegate) JobScheduled(process *eon.Process)                        {}
func (d customDelegate) JobBlocked(process *eon.Process, blockers []*eon.Process) {}
func (d customDelegate) JobStarting(process *eon.Process)                         { _ = fmt.Sprintf("%#v", struct{}{}) }
func (d customDelegate) JobProgressed(process *eon.Process, payload any)          {}
func (d customDelegate) JobEnded(process *eon.Process, err error)                 {}

// -benchtime=4s -count 3
// goos: darwin
// goarch: amd64
// pkg: github.com/hx/eon/events
// cpu: Intel(R) Core(TM) i7-9750H CPU @ 2.60GHz
// BenchmarkDelegation/composed_delegate-12         	32072440	       142.9 ns/op
// BenchmarkDelegation/composed_delegate-12         	33684553	       142.1 ns/op
// BenchmarkDelegation/composed_delegate-12         	33686553	       142.5 ns/op
// BenchmarkDelegation/event_handler-12             	32584875	       147.8 ns/op
// BenchmarkDelegation/event_handler-12             	32552386	       146.7 ns/op
// BenchmarkDelegation/event_handler-12             	33001923	       145.7 ns/op
// BenchmarkDelegation/custom_delegate-12           	34307205	       138.8 ns/op
// BenchmarkDelegation/custom_delegate-12           	34552833	       138.9 ns/op
// BenchmarkDelegation/custom_delegate-12           	34340464	       138.9 ns/op
//
// e.d.: removing calls to handler's mutex RLock and RUnlock made its time virtually
// identical to that of composed_delegate's.
func BenchmarkDelegation(b *testing.B) {
	composed := &events.Delegate{
		Starting: func(process *eon.Process) { _ = fmt.Sprintf("%#v", struct{}{}) },
	}
	handler := new(events.Handler)
	handler.OnStarting(func(process *eon.Process) { _ = fmt.Sprintf("%#v", struct{}{}) })
	for _, delegate := range []struct {
		name string
		d    eon.Delegate
	}{
		{"composed_delegate", composed},
		{"event_handler", handler},
		{"custom_delegate", &customDelegate{}},
	} {
		delegate := delegate
		b.Run(delegate.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				delegate.d.JobStarting(&eon.Process{})
			}
		})
	}
}
