package util

import (
	"container/heap"
	"sync"
)

type blockingHeapBase[T any] struct {
	items []T
	less  func(a, b T) bool
}

func (h *blockingHeapBase[T]) Len() int           { return len(h.items) }
func (h *blockingHeapBase[T]) Less(i, j int) bool { return h.less(h.items[i], h.items[j]) }
func (h *blockingHeapBase[T]) Swap(i, j int)      { h.items[i], h.items[j] = h.items[j], h.items[i] }
func (h *blockingHeapBase[T]) Push(x any)         { h.items = append(h.items, x.(T)) }

func (h *blockingHeapBase[T]) Pop() (item any) {
	index := len(h.items) - 1
	item = h.items[index]
	h.items = h.items[:index]
	return
}

type BlockingHeap[T comparable] struct {
	base *blockingHeapBase[T]
	wg   sync.WaitGroup
}

func NewBlockingHeap[T comparable](less func(a, b T) bool) *BlockingHeap[T] {
	return &BlockingHeap[T]{base: &blockingHeapBase[T]{less: less}}
}

func (h *BlockingHeap[T]) Push(item T) {
	h.wg.Add(1)
	heap.Push(h.base, item)
}

func (h *BlockingHeap[T]) Pop() (item T) {
	item = heap.Pop(h.base).(T)
	h.wg.Done()
	return
}
func (h *BlockingHeap[T]) Len() int { return len(h.base.items) }

func (h *BlockingHeap[T]) Peek() (item T) {
	if index := len(h.base.items) - 1; index >= 0 {
		item = h.base.items[index]
	}
	return
}

func (h *BlockingHeap[T]) Slice() (slice []T) {
	slice = make([]T, len(h.base.items))
	copy(slice, h.base.items)
	return
}

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

func (h *BlockingHeap[T]) Clear() (cleared []T) {
	if len(h.base.items) == 0 {
		return
	}
	cleared = h.base.items
	h.base.items = nil
	h.wg.Add(-len(cleared))
	return
}
