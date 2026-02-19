package function

func SafeCall[T any](fn func(T), param T) {
	if fn != nil {
		fn(param)
	}
}
