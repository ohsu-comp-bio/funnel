package events

import (
	"fmt"
	"github.com/ohsu-comp-bio/funnel/util"
	"regexp"
	"strings"
)

// converts an argument list to a map, e.g.
// ("key", value, "key2", value2) => {"key": value, "key2", value2}
func fields(args ...interface{}) map[string]string {
	ss := make(map[string]string)
	si := make(map[string]interface{})
	si = util.ArgListToMap(args...)
	for k, v := range si {
		ss[k] = fmt.Sprintf("%+v", v)
	}
	return ss
}

// SysLogString returns a flattened string representation of the SystemLog event
func (s *Event) SysLogString() string {
	if s.Type != Type_SYSTEM_LOG {
		return ""
	}
	parts := []string{
		fmt.Sprintf("level='%s'", s.GetSystemLog().Level),
		fmt.Sprintf("msg='%s'", escape(s.GetSystemLog().Msg)),
		fmt.Sprintf("timestamp='%s'", s.Timestamp),
	}
	for k, v := range s.GetSystemLog().Fields {
		parts = append(parts, fmt.Sprintf("%s='%s'", safeKey(k), escape(v)))
	}
	return strings.Join(parts, " ")
}

func escape(s string) string {
	return strings.Replace(s, "'", "\\'", -1)
}

func safeKey(s string) string {
	re := regexp.MustCompile("[\\s]+")
	return re.ReplaceAllString(s, "_")
}
