package events

import (
	"bytes"

	"github.com/golang/protobuf/jsonpb"
)

// Marshaler provides a default JSON marshaler.
var Marshaler = &jsonpb.Marshaler{
	EnumsAsInts:  false,
	EmitDefaults: false,
	Indent:       "\t",
}

// Marshal marshals the event to JSON.
func Marshal(ev *Event) (string, error) {
	return Marshaler.MarshalToString(ev)
}

// Unmarshal unmarshals the event from JSON.
func Unmarshal(b []byte, ev *Event) error {
	r := bytes.NewReader(b)
	return jsonpb.Unmarshal(r, ev)
}
