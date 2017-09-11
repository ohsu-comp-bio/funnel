package sh

import (
	"bytes"
	"github.com/ohsu-comp-bio/funnel/tests/e2e"
	"github.com/sergi/go-diff/diffmatchpatch"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"testing"
)

func TestScripts(t *testing.T) {

	// Find all the test scripts.
	// Start a subtest for each test script.
	ps, _ := filepath.Glob("*.sh")
	for _, p := range ps {
		t.Run(p, func(t *testing.T) {

			conf := e2e.DefaultConfig()
			fun := e2e.NewFunnel(conf)
			fun.WithLocalBackend()
			fun.StartServer()

			var stdout, stderr bytes.Buffer
			cmd := exec.Command("/bin/sh", p)
			cmd.Env = append(os.Environ(), "FUNNEL_SERVER="+conf.Server.HTTPAddress())
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr
			cmd.Run()

			check(t, p+".stdout", stdout.Bytes())
			check(t, p+".stderr", stderr.Bytes())
		})
	}
}

func check(t *testing.T, name string, result []byte) {
	// Load the expected content from a file, if it exists and isn't empty.
	if b, err := ioutil.ReadFile(name); err == nil && len(b) > 0 {
		// Delete id, startTime, endTime, hostIp from task messages
		// because they are non-deterministic
		rx := regexp.MustCompile(`\s+"(id|startTime|endTime|hostIp)": ".*",?`)
		result = rx.ReplaceAll(result, nil)

		if !bytes.Equal(result, b) {
			t.Error(name + " not matched")

			dmp := diffmatchpatch.New()
			d := dmp.DiffMain(string(result), string(b), true)
			d = dmp.DiffCleanupSemantic(d)
			d = dmp.DiffCleanupMerge(d)
			d = dmp.DiffCleanupEfficiency(d)

			t.Logf("diff\n%s", dmp.DiffPrettyText(d))
			//t.Logf("result\n%s", string(result))
			//t.Logf("expected\n%s", string(b))
		}
	} else if len(result) > 0 {
		t.Errorf("missing expected output for %s. got result:\n%s", name, string(result))
	}
}
