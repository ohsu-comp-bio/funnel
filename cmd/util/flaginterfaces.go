package util

import (
	"fmt"
	"time"

	"google.golang.org/protobuf/types/known/durationpb"
)

// DurationValue is a pflag.Value implementation for durationpb.Duration.
type DurationValue struct {
	D **durationpb.Duration // Pointer to a pointer to handle struct fields
}

// String returns the string representation of the duration.
func (d *DurationValue) String() string {
	if d.D == nil || *d.D == nil {
		return ""
	}
	return time.Duration((*d.D).Seconds*1e9 + int64((*d.D).Nanos)).String()
}

// Set parses a string into a durationpb.Duration.
func (d *DurationValue) Set(s string) error {
	dur, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	if d.D == nil {
		return fmt.Errorf("DurationValue.D is nil")
	}
	*d.D = durationpb.New(dur)
	return nil
}

// Type returns the flag type.
func (d *DurationValue) Type() string {
	return "Duration"
}
