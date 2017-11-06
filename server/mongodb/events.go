package mongodb

import (
	"fmt"
	"github.com/ohsu-comp-bio/funnel/events"
	"golang.org/x/net/context"
)

// CreateEvent creates an event for the server to handle.
func (db *MongoDB) CreateEvent(ctx context.Context, req *events.Event) (*events.CreateEventResponse, error) {
	return nil, fmt.Errorf("CreateEvent - Not Implemented")
}

// Write ... TODO
func (db *MongoDB) Write(req *events.Event) error {
	return fmt.Errorf("Write - Not Implemented")
}
