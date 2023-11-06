package mapsx

func MapToSlice[T comparable, U, V any](m map[T]U, f func(T, U) V) []V {
	slice := make([]V, len(m))
	idx := 0
	for k, v := range m {
		slice[idx] = f(k, v)
	}
	return slice
}
