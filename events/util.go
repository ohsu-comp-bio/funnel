package events

import (
	"github.com/golang/protobuf/ptypes"
	tspb "github.com/golang/protobuf/ptypes/timestamp"
	"time"
)

// Timestamp converts a time.Time to a timestamp.
func Timestamp(t time.Time) *tspb.Timestamp {
	p, _ := ptypes.TimestampProto(t)
	return p
}

// TimestampString converts a timestamp to an RFC3339 formatted string.
func TimestampString(t *tspb.Timestamp) string {
	return ptypes.TimestampString(t)
}
