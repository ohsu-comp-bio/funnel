package util

import (
	"errors"
	"strings"
)

// MultiError helps collect multiple errors and implements Go's "error" interface.
type MultiError []error

// Error returns all the error strings joined by a newline.
func (m MultiError) Error() string {
	var strs []string
	for _, e := range m {
		if e != nil {
			strs = append(strs, e.Error())
		}
	}
	return strings.Join(strs, "\n")
}

// IsNil returns true if all errors in the slice are nil.
func (m MultiError) IsNil() bool {
	isNil := true
	for _, e := range m {
		if e != nil {
			isNil = false
		}
	}
	return isNil
}

// ToError returns an error interface.
func (m MultiError) ToError() error {
	if m.IsNil() {
		return nil
	}
	return errors.New(m.Error())
}
