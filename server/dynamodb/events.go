package dynamodb

import (
	"fmt"
	"github.com/ohsu-comp-bio/funnel/events"
	"golang.org/x/net/context"
)

// CreateEvent creates an event for the server to handle.
func (db *DynamoDB) CreateEvent(ctx context.Context, req *events.Event) (*events.CreateEventResponse, error) {
	return nil, fmt.Errorf("CreateEvent - Not Implemented")
}
