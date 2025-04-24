package util

import (
	"time"

	"google.golang.org/protobuf/types/known/durationpb"
)

type DurationValue struct {
	D *durationpb.Duration
}

func (d *DurationValue) String() string {
	if d.D == nil {
		return ""
	}
	return time.Duration(d.D.Seconds*1e9 + int64(d.D.Nanos)).String()
}

func (d *DurationValue) Set(s string) error {
	dur, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	d.D = durationpb.New(dur)
	return nil
}

func (*DurationValue) Type() string {
	return "Duration"
}
