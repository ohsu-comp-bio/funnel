package events

import (
	"bytes"
	"github.com/golang/protobuf/jsonpb"
)

var Marshaler = &jsonpb.Marshaler{
	EnumsAsInts:  false,
	EmitDefaults: false,
	Indent:       "\t",
}

func Marshal(ev *Event) (string, error) {
	return Marshaler.MarshalToString(ev)
}

func Unmarshal(b []byte, ev *Event) error {
	r := bytes.NewReader(b)
	return jsonpb.Unmarshal(r, ev)
}
