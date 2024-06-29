package lang

func NewPtr[T any](value T) *T {
	return &value
}
