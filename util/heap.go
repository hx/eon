package util

import "container/heap"

type heapBase[T any] struct {
	items []T
	less  func(a, b T) bool
}

func (h *heapBase[T]) Len() int           { return len(h.items) }
func (h *heapBase[T]) Less(i, j int) bool { return h.less(h.items[i], h.items[j]) }
func (h *heapBase[T]) Swap(i, j int)      { h.items[i], h.items[j] = h.items[j], h.items[i] }
func (h *heapBase[T]) Push(x any)         { h.items = append(h.items, x.(T)) }

func (h *heapBase[T]) Pop() (item any) {
	index := len(h.items) - 1
	item = h.items[index]
	h.items = h.items[:index]
	return
}

type Heap[T comparable] struct {
	base *heapBase[T]
}

func NewHeap[T comparable](less func(a, b T) bool) *Heap[T] {
	return &Heap[T]{base: &heapBase[T]{less: less}}
}

func (h *Heap[T]) Push(item T)   { heap.Push(h.base, item) }
func (h *Heap[T]) Pop() (item T) { return heap.Pop(h.base).(T) }
func (h *Heap[T]) Len() int      { return len(h.base.items) }

func (h *Heap[T]) Peek() (item T) {
	if index := len(h.base.items) - 1; index >= 0 {
		item = h.base.items[index]
	}
	return
}

func (h *Heap[T]) Slice() (slice []T) {
	slice = make([]T, len(h.base.items))
	copy(slice, h.base.items)
	return
}

func (h *Heap[T]) Remove(item T) bool {
	// TODO: consider optimising with a map
	for i, v := range h.base.items {
		if v == item {
			heap.Remove(h.base, i)
			return true
		}
	}
	return false
}

func (h *Heap[T]) Clear() (cleared []T) {
	if len(h.base.items) == 0 {
		return
	}
	cleared = h.base.items
	h.base.items = nil
	return
}
