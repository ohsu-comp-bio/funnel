// Package examples bundles example tasks into the Funnel CLI.
package examples

import (
	"path"
	"strings"

	intern "github.com/ohsu-comp-bio/funnel/examples/internal"
)

var examples = buildExamples()

func buildExamples() map[string]string {
	examples := map[string]string{}
	for _, n := range intern.AssetNames() {
		sn := path.Base(n)
		sn = strings.TrimSuffix(sn, path.Ext(sn))
		b := intern.MustAsset(n)
		examples[sn] = string(b)
	}
	return examples
}

// Examples returns a set of example tasks.
func Examples() map[string]string {
	return examples
}
