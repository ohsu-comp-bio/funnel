package run

import (
	"os"
	"testing"

	"github.com/go-test/deep"
	"github.com/ohsu-comp-bio/funnel/tes"
)

func TestParse(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	expected := []*tes.Task{
		{
			Name:        "foo",
			Description: "mydesc",
			Inputs: []*tes.Input{
				{
					Name: "f1",
					Url:  "file://" + cwd + "/testdata/f1.txt",
					Path: "/opt/funnel/inputs" + cwd + "/testdata/f1.txt",
				},
				{
					Name: "f2",
					Url:  "file://" + cwd + "/testdata/f2.txt",
					Path: "/opt/funnel/inputs" + cwd + "/testdata/f2.txt",
				},
				{
					Name: "f4",
					Url:  "file://" + cwd + "/testdata/f4",
					Path: "/opt/funnel/inputs" + cwd + "/testdata/f4",
					Type: tes.FileType_DIRECTORY,
				},
				{
					Name:    "c1",
					Path:    "/opt/funnel/inputs" + cwd + "/testdata/content.txt",
					Content: "test content\n",
				},
			},
			Outputs: []*tes.Output{
				{
					Name: "stdout-0",
					Url:  "file://" + cwd + "/testdata/stdout-first",
					Path: "/opt/funnel/outputs/stdout-0",
				},
				{
					Name: "stdout-1",
					Url:  "file://" + cwd + "/testdata/stdout-second",
					Path: "/opt/funnel/outputs/stdout-1",
				},
				{
					Name: "stderr-1",
					Url:  "file://" + cwd + "/testdata/stderr-second",
					Path: "/opt/funnel/outputs/stderr-1",
				},
				{
					Name: "f3",
					Url:  "file://" + cwd + "/testdata/f3",
					Path: "/opt/funnel/outputs" + cwd + "/testdata/f3",
				},
				{
					Name: "o9",
					Url:  "file://" + cwd + "/testdata/o9",
					Path: "/opt/funnel/outputs" + cwd + "/testdata/o9",
					Type: tes.FileType_DIRECTORY,
				},
			},
			Resources: &tes.Resources{
				CpuCores:    8,
				Preemptible: true,
				RamGb:       32.0,
				DiskGb:      100.0,
				Zones:       []string{"zone1", "zone2"},
			},
			Executors: []*tes.Executor{
				{
					Image:   "busybox",
					Command: []string{"sh", "-c", "echo hello"},
					Workdir: "myworkdir",
					Stdout:  "/opt/funnel/outputs/stdout-0",
					Stderr:  "/opt/funnel/outputs/stderr-0",
					Env: map[string]string{
						"c1": "/opt/funnel/inputs" + cwd + "/testdata/content.txt",
						"e1": "e1v",
						"e2": "e2v",
						"f1": "/opt/funnel/inputs" + cwd + "/testdata/f1.txt",
						"f2": "/opt/funnel/inputs" + cwd + "/testdata/f2.txt",
						"f4": "/opt/funnel/inputs" + cwd + "/testdata/f4",
						"f3": "/opt/funnel/outputs" + cwd + "/testdata/f3",
						"o9": "/opt/funnel/outputs" + cwd + "/testdata/o9",
					},
				},
				{
					Image:   "busybox",
					Command: []string{"echo", "two"},
					Workdir: "myworkdir",
					Stdout:  "/opt/funnel/outputs/stdout-1",
					Stderr:  "/opt/funnel/outputs/stderr-1",
					Env: map[string]string{
						"c1": "/opt/funnel/inputs" + cwd + "/testdata/content.txt",
						"e1": "e1v",
						"e2": "e2v",
						"f1": "/opt/funnel/inputs" + cwd + "/testdata/f1.txt",
						"f2": "/opt/funnel/inputs" + cwd + "/testdata/f2.txt",
						"f4": "/opt/funnel/inputs" + cwd + "/testdata/f4",
						"f3": "/opt/funnel/outputs" + cwd + "/testdata/f3",
						"o9": "/opt/funnel/outputs" + cwd + "/testdata/o9",
					},
				},
			},
			Volumes: []string{"/volone", "/voltwo"},
			Tags: map[string]string{
				"one": "onev",
				"two": "twov",
			},
		},
	}

	result, perr := ParseString(`
    'echo hello'
    --container busybox
    --name myname
    --description mydesc
    --tag one=onev
    --tag two=twov
    --in f1=./testdata/f1.txt
    -i f2=./testdata/f2.txt
    -o f3=./testdata/f3
    -I f4=./testdata/f4
    -e e1=e1v
    --env e2=e2v
    --stdout ./testdata/stdout-first
    -x '--name foo'
    --exec 'echo two'
    --stdout ./testdata/stdout-second
    --vol /volone
    --vol /voltwo
    --cpu 8
    --ram 32
    --disk 100
    --preemptible
    --zone zone1
    --zone zone2
    -O o9=./testdata/o9
    --stderr ./testdata/stderr-second
    -S http://localhost:9001
    -p
    -w myworkdir
    -C c1=./testdata/content.txt
  `)

	if perr != nil {
		t.Fatal(perr)
	}

	if diff := deep.Equal(result, expected); diff != nil {
		s, _ := tes.MarshalToString(expected[0])
		t.Log("Expected", s)
		q, _ := tes.MarshalToString(result[0])
		t.Log("Actual", q)
		for _, d := range diff {
			t.Log("Diff", d)
		}
		t.Fatal("unexpected results")
	}
}
