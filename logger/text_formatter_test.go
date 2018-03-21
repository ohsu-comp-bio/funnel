package logger

import (
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/ohsu-comp-bio/funnel/tes"
)

func TestFormatNilProtoField(t *testing.T) {
	if r := recover(); r != nil {
		t.Fatal("error")
	}
	var nt *tes.Task

	c := DebugConfig()
	tf := &textFormatter{
		c.TextFormat,
		jsonFormatter{
			conf: c.JSONFormat,
		},
	}

	entry := logrus.WithFields(logrus.Fields{
		"ns":        "TEST",
		"nil value": nt,
	})
	tf.Format(entry)
}
