package repository

// ErrNotFound is a custom error type for not-found conditions.
type ErrNotFound string

func (e ErrNotFound) Error() string {
	return string(e)
}