package logger

import (
	"bytes"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/golang/protobuf/proto"
	"github.com/kr/pretty"
	"github.com/logrusorgru/aurora"
	"io"
	"runtime"
	"sort"
	"time"
)

var baseTimestamp = time.Now()

type textFormatter struct {
	TextFormatConfig
	json jsonFormatter
}

func isColorTerminal(w io.Writer) bool {
	return logrus.IsTerminal(w) && (runtime.GOOS != "windows")
}

func (f *textFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	isColored := (f.ForceColors || isColorTerminal(entry.Logger.Out)) && !f.DisableColors
	if !isColored {
		return f.json.Format(entry)
	}

	// entry namespace
	ns := entry.Data["ns"].(string)

	// Gather keys so they can be sorted
	keys := make([]string, 0, len(entry.Data))
	for k := range entry.Data {
		// "ns" (namespace) always comes first, so skip that one.
		if k != "ns" {
			keys = append(keys, k)
		}
	}

	if !f.DisableSorting {
		sort.Strings(keys)
	}

	b := entry.Buffer
	if b == nil {
		b = &bytes.Buffer{}
	}

	prefixFieldClashes(entry.Data)

	if !f.DisableTimestamp {
		if !f.FullTimestamp {
			// How many seconds since this package was initialized
			t := entry.Time.Sub(baseTimestamp) / time.Second
			entry.Data["time"] = fmt.Sprintf("%04d", int(t))
		} else {
			entry.Data["time"] = entry.Time.Format(f.TimestampFormat)
		}
	}

	if entry.Message != "" {
		entry.Data["msg"] = entry.Message
	}

	var levelColor aurora.Color

	switch entry.Level {
	case logrus.DebugLevel:
		levelColor = aurora.BrownFg
	case logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel:
		levelColor = aurora.RedFg
	default:
		levelColor = aurora.CyanFg
	}
	nsColor := levelColor | aurora.BoldFm

	fmt.Fprintf(b, "%-20s %s\n", aurora.Colorize(ns, nsColor), entry.Message)

	for _, k := range keys {
		v := entry.Data[k]

		switch x := v.(type) {
		case string:
		case int:
		case int8:
		case int16:
		case int32:
		case int64:
		case uint8:
		case uint16:
		case uint32:
		case uint64:
		case complex64:
		case complex128:
		case float32:
		case float64:
		case bool:
		case proto.Message:
			v = pretty.Sprint(x)
		case fmt.Stringer:
		case error:
		default:
			v = pretty.Sprint(x)
		}
		fmt.Fprintf(b, "%-20s %v\n", aurora.Colorize(k, levelColor), v)
	}

	b.WriteByte('\n')
	return b.Bytes(), nil
}

// This is to not silently overwrite `time`, `msg` and `level` fields when
// dumping it. If this code wasn't there doing:
//
//  logrus.WithField("level", 1).Info("hello")
//
// Would just silently drop the user provided level. Instead with this code
// it'll logged as:
//
//  {"level": "info", "fields.level": 1, "msg": "hello", "time": "..."}
//
// It's not exported because it's still using Data in an opinionated way. It's to
// avoid code duplication between the two default formatters.
func prefixFieldClashes(data logrus.Fields) {
	if t, ok := data["time"]; ok {
		data["fields.time"] = t
	}

	if m, ok := data["msg"]; ok {
		data["fields.msg"] = m
	}

	if l, ok := data["level"]; ok {
		data["fields.level"] = l
	}
}
