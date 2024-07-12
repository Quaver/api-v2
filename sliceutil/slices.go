package sliceutil

// All Checks if all elements in a slice match a given predicate
func All[T any](ts []T, pred func(T) bool) bool {
	for _, t := range ts {
		if !pred(t) {
			return false
		}
	}
	return true
}

// Values Converts a map's values to a slice
func Values[M ~map[K]V, K comparable, V any](m M) []V {
	r := make([]V, 0, len(m))
	for _, v := range m {
		r = append(r, v)
	}
	return r
}

// Keys Converts a map's keys into a slice
func Keys[M ~map[K]V, K comparable, V any](m M) []K {
	r := make([]K, 0, len(m))
	for k, _ := range m {
		r = append(r, k)
	}
	return r
}
