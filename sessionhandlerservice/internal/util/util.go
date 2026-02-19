package util

func Ptr[T any](v T) *T {
	return &v
}

func ArrayInitializer[T any](slice []T) []T {
	if len(slice) == 0 {
		return make([]T, 0)
	}

	return slice
}
