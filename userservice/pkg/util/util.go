package util

func Ptr[T any](v T) *T {
	return &v
}

func UpdateIfNotNil[T any](target *T, source *T) {
	if source != nil {
		*target = *source
	}
}
