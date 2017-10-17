package util

// ErrResourceExists is an error used for communicating when a resource exists
// and does not need to be created
type ErrResourceExists struct{}

func (e ErrResourceExists) Error() string {
	return "resource exists"
}
