package scheduler

import (
	"fmt"
	"github.com/go-test/deep"
	"github.com/ohsu-comp-bio/funnel/config"
	pbf "github.com/ohsu-comp-bio/funnel/proto/funnel"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"io/ioutil"
	"testing"
)

func TestScheduleSingleTaskWorker(t *testing.T) {
	// Test that task resources get set in worker
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

	result := ScheduleSingleTaskWorker(pre, c.Worker, task)

	expected := &Offer{
		TaskID: "task1",
		Worker: &pbf.Worker{
			Id: "test-prefix-task1",
			Resources: &pbf.Resources{
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

	// Test that worker config resources override task resources get set in worker
	c.Worker.Resources = config.Resources{
		Cpus:   2,
		RamGb:  8.0,
		DiskGb: 100.0,
	}

	result = ScheduleSingleTaskWorker(pre, c.Worker, task)

	expected = &Offer{
		TaskID: "task1",
		Worker: &pbf.Worker{
			Id: "test-prefix-task1",
			Resources: &pbf.Resources{
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

func TestSetupTemplatedHPCWorker(t *testing.T) {
	tmp, err := ioutil.TempDir("", "funnel-test-scheduler")
	if err != nil {
		t.Fatal(err)
	}

	c := config.DefaultConfig()
	c.Worker.WorkDir = tmp

	w := &pbf.Worker{
		Id: "test-worker",
		Resources: &pbf.Resources{
			Cpus:   1,
			RamGb:  1.0,
			DiskGb: 10.0,
		},
	}

	tpl := `
#!/bin/bash
#TEST --name {{.WorkerId}}
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

{{.Executable}} worker start --config {{.WorkerConfig}}
`

	sf, err := SetupTemplatedHPCWorker("slurm-test", tpl, c, w)
	if err != nil {
		t.Fatal(err)
	}

	actual, rerr := ioutil.ReadFile(sf)
	if rerr != nil {
		t.Fatal(rerr)
	}

	workerPath, err := DetectWorkerPath()
	if err != nil {
		t.Fatal(err)
	}

	expected := `
#!/bin/bash
#TEST --name test-worker
#TEST --flag
#TEST -e %s/test-worker/stderr
#TEST -o %s/test-worker/stdout
#TEST --cpus 1
#TEST --mem 1GB
#TEST --disk 10.0GB

%s worker start --config %s/test-worker/worker.conf.yml
`

	expected = fmt.Sprintf(expected, tmp, tmp, workerPath, tmp)

	if string(actual) != expected {
		log.Error("Expected", "", expected)
		log.Error("Actual", "", string(actual))
		t.Fatal("Unexpected content")
	}
}
