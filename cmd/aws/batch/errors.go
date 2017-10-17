package batch

type errResourceExists struct{}

func (e errResourceExists) Error() string {
	return "resource exists"
}
