package util

import (
	"github.com/spf13/pflag"
  "github.com/ohsu-comp-bio/funnel/config"
)

func ConfigFlags(conf *config.Config) *pflag.FlagSet {
	f := pflag.NewFlagSet("", pflag.ContinueOnError)

  {{ range .Leaves -}}
    {{ if .IsValueType }}
      f.Var(&conf.{{ join .Key }}, "{{ join .Key }}", "{{ synopsis .Doc }}")
    {{ else }}
      f.{{ pflagType . }}Var(&conf.{{ join .Key }}, "{{ join .Key }}", conf.{{ join .Key }}, "{{ synopsis .Doc }}")
    {{ end }}
  {{ end }}

  return f
}
