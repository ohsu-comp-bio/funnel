package scheduler

import (
	"fmt"
	"github.com/go-test/deep"
	"github.com/ohsu-comp-bio/funnel/config"
	pbs "github.com/ohsu-comp-bio/funnel/proto/scheduler"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"io/ioutil"
	"testing"
)

func TestSetupSingleTaskNode(t *testing.T) {
	// Test that task resources get set in node
	pre := "test-prefix-"
	c := config.DefaultConfig()
	task := &tes.Task{
		Id: "task1",
		Resources: &tes.Resources{
			CpuCores: 1,
			RamGb:    4.0,
			SizeGb:   10.0,
		},
	}

	result := SetupSingleTaskNode(pre, c.Scheduler.Node, task)

	expected := &Offer{
		TaskID: "task1",
		Node: &pbs.Node{
			Id: "test-prefix-task1",
			Resources: &pbs.Resources{
				Cpus:   1,
				RamGb:  4.0,
				DiskGb: 10.0,
			},
			Metadata: map[string]string{"project": ""},
		},
		Scores: Scores{},
	}

	if diff := deep.Equal(result, expected); diff != nil {
		log.Debug("Expected", fmt.Sprintf("%+v", expected))
		log.Debug("Actual", fmt.Sprintf("%+v", result))
		for _, d := range diff {
			log.Debug("Diff", d)
		}
		t.Fatal("unexpected Offer")
	}

	// Test that node config resources override task resources get set in node
	c.Scheduler.Node.Resources.Cpus = 2
	c.Scheduler.Node.Resources.RamGb = 8.0
	c.Scheduler.Node.Resources.DiskGb = 100.0

	result = SetupSingleTaskNode(pre, c.Scheduler.Node, task)

	expected = &Offer{
		TaskID: "task1",
		Node: &pbs.Node{
			Id: "test-prefix-task1",
			Resources: &pbs.Resources{
				Cpus:   2,
				RamGb:  8.0,
				DiskGb: 100.0,
			},
			Metadata: map[string]string{"project": ""},
		},
		Scores: Scores{},
	}

	if diff := deep.Equal(result, expected); diff != nil {
		log.Debug("Expected", fmt.Sprintf("%+v", expected))
		log.Debug("Actual", fmt.Sprintf("%+v", result))
		for _, d := range diff {
			log.Debug("Diff", d)
		}
		t.Fatal("unexpected Offer")
	}
}

func TestSetupTemplatedHPCNode(t *testing.T) {
	tmp, err := ioutil.TempDir("", "funnel-test-scheduler")
	if err != nil {
		t.Fatal(err)
	}

	c := config.DefaultConfig()
	c.Scheduler.Node.WorkDir = tmp

	n := &pbs.Node{
		Id: "test-node",
		Resources: &pbs.Resources{
			Cpus:   1,
			RamGb:  1.0,
			DiskGb: 10.0,
		},
	}

	tpl := `
#!/bin/bash
#TEST --name {{.NodeId}}
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

{{.Executable}} node run --config {{.Config}}
`

	sf, err := SetupTemplatedHPCNode("slurm-test", tpl, c, n)
	if err != nil {
		t.Fatal(err)
	}

	actual, rerr := ioutil.ReadFile(sf)
	if rerr != nil {
		t.Fatal(rerr)
	}

	binaryPath, err := DetectBinaryPath()
	if err != nil {
		t.Fatal(err)
	}

	expected := `
#!/bin/bash
#TEST --name test-node
#TEST --flag
#TEST -e %s/test-node/stderr
#TEST -o %s/test-node/stdout
#TEST --cpus 1
#TEST --mem 1GB
#TEST --disk 10.0GB

%s node run --config %s/test-node/node.conf.yml
`

	expected = fmt.Sprintf(expected, tmp, tmp, binaryPath, tmp)

	if string(actual) != expected {
		log.Error("Expected", "", expected)
		log.Error("Actual", "", string(actual))
		t.Fatal("Unexpected content")
	}
}
