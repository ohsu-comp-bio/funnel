package util

import (
  "github.com/rs/xid"
)

// GenTaskID generates a task ID string.
// IDs are globally unique and sortable.
func GenTaskID() string {
	id := xid.New()
	return id.String()
}
