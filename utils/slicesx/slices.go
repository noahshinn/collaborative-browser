package slicesx

func Map[T, U any](s []T, f func(T) U) []U {
	mapped := make([]U, len(s))
	for i, v := range s {
		mapped[i] = f(v)
	}
	return mapped
}

func Filter[T any](s []T, f func(T) bool) []T {
	filtered := make([]T, 0)
	for _, v := range s {
		if f(v) {
			filtered = append(filtered, v)
		}
	}
	return filtered
}

func Any[T any](s []T, f func(T) bool) bool {
	for _, v := range s {
		if f(v) {
			return true
		}
	}
	return false
}

func All[T any](s []T, f func(T) bool) bool {
	for _, v := range s {
		if !f(v) {
			return false
		}
	}
	return true
}

func Contains[T comparable](s []T, v T) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}
