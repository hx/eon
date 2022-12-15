package util

type Set[T comparable] map[T]struct{}

func NewSet[T comparable](items ...T) (result Set[T]) {
	result = make(Set[T], len(items))
	for _, item := range items {
		result.Add(item)
	}
	return
}

func (s Set[T]) Add(item T)    { s[item] = struct{}{} }
func (s Set[T]) Remove(item T) { delete(s, item) }

func (s Set[T]) Contains(item T) (found bool) {
	_, found = s[item]
	return
}

func (s Set[T]) Slice() (result []T) {
	result = make([]T, len(s))
	var i int
	for item := range s {
		result[i] = item
		i++
	}
	return
}
