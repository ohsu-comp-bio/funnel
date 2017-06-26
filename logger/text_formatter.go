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

	b := entry.Buffer
	if b == nil {
		b = &bytes.Buffer{}
	}

	if !f.DisableTimestamp {
		if !f.FullTimestamp {
			// How many seconds since this package was initialized
			t := entry.Time.Sub(baseTimestamp) / time.Second
			entry.Data["time"] = fmt.Sprintf("%04d", int(t))
		} else {
			entry.Data["time"] = entry.Time.Format(f.TimestampFormat)
		}
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

	for _, k := range f.entryDataKeys(entry) {
		v := entry.Data[k]

		// Some keys can conflict with reserved names.
		// If they do, resolve the conflict with a prefix.
		keyname := k
		switch k {
		case "time", "msg", "level":
			keyname = "fields." + keyname
		}

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

func (f *textFormatter) entryDataKeys(entry *logrus.Entry) []string {

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
	return keys
}
