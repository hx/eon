package util

// Set is a generic set implementation backed by a map.
//
// A zero Set, like a zero map, is not valid. Use NewSet or make(Set[T]) instead.
type Set[T comparable] map[T]struct{}

// NewSet creates a new Set containing the given items.
func NewSet[T comparable](items ...T) (result Set[T]) {
	result = make(Set[T], len(items))
	for _, item := range items {
		result.Add(item)
	}
	return
}

// Add adds the given item to the receiver, if it is not already a member.
func (s Set[T]) Add(item T) { s[item] = struct{}{} }

// Remove removes the given item from the receiver, if it is a member.
func (s Set[T]) Remove(item T) { delete(s, item) }

// Contains returns true if the given item is a member of the receiver.
func (s Set[T]) Contains(item T) (found bool) {
	_, found = s[item]
	return
}

// Slice returns members of the receiver in a slice, in unspecified order.
func (s Set[T]) Slice() (result []T) {
	result = make([]T, len(s))
	var i int
	for item := range s {
		result[i] = item
		i++
	}
	return
}
