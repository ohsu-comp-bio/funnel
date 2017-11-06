package util

import (
	"strings"
)

// MultiError helps collect multiple errors and implements Go's "error" interface.
type MultiError []error

// Error returns all the error strings joined by a newline.
func (m MultiError) Error() string {
	var strs []string
	for _, e := range m {
		strs = append(strs, e.Error())
	}
	return strings.Join(strs, "\n")
}
