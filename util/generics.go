package util

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
