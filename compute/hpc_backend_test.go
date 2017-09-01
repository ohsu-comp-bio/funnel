package compute

import (
	"fmt"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"io/ioutil"
	"testing"
)

func TestSetupTemplatedHPCSubmit(t *testing.T) {
	tmp, err := ioutil.TempDir("", "funnel-test-scheduler")
	if err != nil {
		t.Fatal(err)
	}

	conf := config.DefaultConfig()
	conf.Worker.WorkDir = tmp

	task := &tes.Task{
		Id: "test-taskid",
		Executors: []*tes.Executor{
			{
				Cmd: []string{"echo test"},
			},
		},
		Resources: &tes.Resources{
			CpuCores: 1,
			RamGb:    1.0,
			SizeGb:   10.0,
		},
	}

	tpl := `
#!/bin/bash
#TEST --name {{.TaskId}}
#TEST --flag
#TEST -e {{.WorkDir}}/stderr
#TEST -o {{.WorkDir}}/stdout
{{if ne .Cpus 0 -}}
{{printf "#TEST --cpus %d" .Cpus}}
{{- end}}
{{if ne .RamGb 0.0 -}}
{{printf "#TEST --mem %.0fGB" .RamGb}}
{{- end}}
{{if ne .DiskGb 0.0 -}}
{{printf "#TEST --disk %.1fGB" .DiskGb}}
{{- end}}

{{.Executable}} worker run --config {{.Config}}
`

	b := HPCBackend{"test", "qsub", conf, tpl}

	sf, err := b.setupTemplatedHPCSubmit(task)
	if err != nil {
		t.Fatal(err)
	}

	actual, rerr := ioutil.ReadFile(sf)
	if rerr != nil {
		t.Fatal(rerr)
	}

	binaryPath, err := DetectFunnelBinaryPath()
	if err != nil {
		t.Fatal(err)
	}

	expected := `
#!/bin/bash
#TEST --name test-taskid
#TEST --flag
#TEST -e %s/test-taskid/stderr
#TEST -o %s/test-taskid/stdout
#TEST --cpus 1
#TEST --mem 1GB
#TEST --disk 10.0GB

%s worker run --config %s/test-taskid/worker.conf.yml
`

	expected = fmt.Sprintf(expected, tmp, tmp, binaryPath, tmp)

	if string(actual) != expected {
		log.Error("Expected", "", expected)
		log.Error("Actual", "", string(actual))
		t.Fatal("Unexpected content")
	}
}
