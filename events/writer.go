package events

import (
	"fmt"
	"github.com/ohsu-comp-bio/funnel/config"
)

// Writer provides write access to a task's events
type Writer interface {
	Write(*Event) error
}

type multiwriter []Writer

// MultiWriter writes events to all the given writers.
func MultiWriter(ws ...Writer) Writer {
	return multiwriter(ws)
}

// Write writes an event to all the writers.
func (mw multiwriter) Write(ev *Event) error {
	for _, w := range mw {
		err := w.Write(ev)
		if err != nil {
			return err
		}
	}
	return nil
}

type discard struct{}

func (discard) Write(*Event) error {
	return nil
}

// Discard is a writer which discards all events.
var Discard = discard{}

// FromConfig returns a Writer based on the given config.
func FromConfig(conf config.EventWriters) (Writer, error) {
	var writers []Writer
	for _, w := range conf.Active {

		var writer Writer
		var err error

		switch w {
		case "dynamodb":
			writer, err = NewDynamoDBEventWriter(conf.DynamoDB)
		case "log":
			writer = NewLogger("worker")
		case "rpc":
			writer, err = NewRPCWriter(conf.RPC)
    case "elastic":
      writer, err = NewElasticWriter(conf.Elastic)
		default:
			err = fmt.Errorf("unknown EventWriter")
		}
		if err != nil {
			return nil, fmt.Errorf("failed to instantiate EventWriter: %v", err)
		}

		writers = append(writers, writer)
	}
	if writers == nil {
		return Discard, nil
	}
	return MultiWriter(writers...), nil
}
