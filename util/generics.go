package util

// Remove removes the given item from the given slice, or returns false if it's not found.
func Remove[T comparable](slice *[]T, item T) (removed bool) {
	old := *slice
	for i, v := range old {
		if v == item {
			*slice = append(old[:i], old[i+1:]...)
			return true
		}
	}
	return
}
