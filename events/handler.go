package events

import (
	"github.com/hx/eon"
	"sync"
)

type handlerFuncs[T any] struct {
	funcs []*T
	mutex sync.RWMutex
}

// Handler is a midground.Delegate implementation that allows discreet event handlers to be added and removed at any
// time.
//
// Handler functions are run in the order in which they were added.
//
// A pointer to a zero Handler is ready to use.
type Handler struct {
	scheduled  handlerFuncs[func(process *eon.Process)]
	blocked    handlerFuncs[func(process *eon.Process, blockers []*eon.Process)]
	starting   handlerFuncs[func(process *eon.Process)]
	progressed handlerFuncs[func(process *eon.Process, payload any)]
	ended      handlerFuncs[func(process *eon.Process, err error)]
}

func addHandler[T any](handler T, collection *handlerFuncs[T]) (remove func()) {
	pointer := &handler
	collection.mutex.Lock()
	collection.funcs = append(collection.funcs, pointer)
	collection.mutex.Unlock()
	return func() {
		collection.mutex.Lock()
		for i, v := range collection.funcs {
			if v == pointer {
				collection.funcs = append(collection.funcs[:i], collection.funcs[i+1:]...)
				break
			}
		}
		collection.mutex.Unlock()
	}
}

// OnScheduled adds a handler for when a midground.Process is scheduled. The returned remove function can be used to
// remove the handler.
func (h *Handler) OnScheduled(fn func(process *eon.Process)) (remove func()) {
	return addHandler(fn, &h.scheduled)
}

// OnBlocked adds a handler for when a midground.Process is blocked. The returned remove function can be used to
// remove the handler.
func (h *Handler) OnBlocked(fn func(process *eon.Process, blockers []*eon.Process)) (remove func()) {
	return addHandler(fn, &h.blocked)
}

// OnStarting adds a handler for when a midground.Process is starting. The returned remove function can be used to
// remove the handler.
func (h *Handler) OnStarting(fn func(process *eon.Process)) (remove func()) {
	return addHandler(fn, &h.starting)
}

// OnProgressed adds a handler for when a midground.Process progresses. The returned remove function can be used to
// remove the handler.
func (h *Handler) OnProgressed(fn func(process *eon.Process, payload any)) (remove func()) {
	return addHandler(fn, &h.progressed)
}

// OnEnded adds a handler for when a midground.Process ends. The returned remove function can be used to
// remove the handler.
func (h *Handler) OnEnded(fn func(process *eon.Process, err error)) (remove func()) {
	return addHandler(fn, &h.ended)
}

func (h *Handler) JobScheduled(process *eon.Process) {
	h.scheduled.mutex.RLock()
	for _, fn := range h.scheduled.funcs {
		(*fn)(process)
	}
	h.scheduled.mutex.RUnlock()
}

func (h *Handler) JobBlocked(process *eon.Process, blockers []*eon.Process) {
	h.blocked.mutex.RLock()
	for _, fn := range h.blocked.funcs {
		(*fn)(process, blockers)
	}
	h.blocked.mutex.RUnlock()
}

func (h *Handler) JobStarting(process *eon.Process) {
	h.starting.mutex.RLock()
	for _, fn := range h.starting.funcs {
		(*fn)(process)
	}
	h.starting.mutex.RUnlock()
}

func (h *Handler) JobProgressed(process *eon.Process, payload any) {
	h.progressed.mutex.RLock()
	for _, fn := range h.progressed.funcs {
		(*fn)(process, payload)
	}
	h.progressed.mutex.RUnlock()
}

func (h *Handler) JobEnded(process *eon.Process, err error) {
	h.ended.mutex.RLock()
	for _, fn := range h.ended.funcs {
		(*fn)(process, err)
	}
	h.ended.mutex.RUnlock()
}
