package util

import (
	"container/heap"
	"sync"
)

// BlockingHeap implements a min-heap for values of type T. It is identical to Heap, but also maintains a sync.WaitGroup
// that unblocks when the heap has no members.
//
// This type isn't being used at the moment, but could be used to replace midground.Scheduler.running if we need to wait
// for all its running jobs to complete.
//
// A zero BlockingHeap is not valid. Use NewBlockingHeap instead.
type BlockingHeap[T comparable] struct {
	wg   sync.WaitGroup
	base *heapBase[T]
}

// NewBlockingHeap creates a new BlockingHeap using the given less function for sort comparisons.
func NewBlockingHeap[T comparable](less func(a, b T) bool) *BlockingHeap[T] {
	return &BlockingHeap[T]{base: &heapBase[T]{less: less}}
}

// Push adds an item to the receiver.
func (h *BlockingHeap[T]) Push(item T) {
	h.wg.Add(1)
	heap.Push(h.base, item)
}

// Pop removes and returns the item at the top of the receiver.
func (h *BlockingHeap[T]) Pop() (item T) {
	item = heap.Pop(h.base).(T)
	h.wg.Done()
	return
}

// Len returns the number of items in the receiver.
func (h *BlockingHeap[T]) Len() int { return len(h.base.items) }

// Peek returns the item at the top of the receiver without removing it. If the receiver is empty, a zero value of the
// receiver's type T is returned instead.
func (h *BlockingHeap[T]) Peek() (item T) {
	if index := len(h.base.items) - 1; index >= 0 {
		item = h.base.items[index]
	}
	return
}

// Slice returns a slice of all items in the receiver. The item at the top of the receiver will be at the end of the
// slice. The order of all other items should be considered random.
func (h *BlockingHeap[T]) Slice() (slice []T) {
	slice = make([]T, len(h.base.items))
	copy(slice, h.base.items)
	return
}

// Remove removes the given item from the receiver.
func (h *BlockingHeap[T]) Remove(item T) bool {
	// TODO: consider optimising with a map
	for i, v := range h.base.items {
		if v == item {
			heap.Remove(h.base, i)
			h.wg.Done()
			return true
		}
	}
	return false
}

// Clear removes all items from the receiver.
func (h *BlockingHeap[T]) Clear() (cleared []T) {
	if len(h.base.items) == 0 {
		return
	}
	cleared = h.base.items
	h.base.items = nil
	h.wg.Add(-len(cleared))
	return
}

// Wait blocks while the receiver is not empty.
func (h *BlockingHeap[T]) Wait() { h.wg.Wait() }
