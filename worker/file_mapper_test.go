package worker

import (
	"fmt"
	"github.com/go-test/deep"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"io/ioutil"
	"os"
	"testing"
)

func TestMapTask(t *testing.T) {
	tmp, err := ioutil.TempDir("", "funnel-test-mapper")
	if err != nil {
		t.Fatal(err)
	}
	f := FileMapper{
		dir: tmp,
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	task := &tes.Task{
		Inputs: []*tes.TaskParameter{
			{
				Name: "f1",
				Url:  "file://" + cwd + "/testdata/f1.txt",
				Path: "/opt/funnel/inputs/testdata/f1.txt",
			},
			{
				Name: "f4",
				Url:  "file://" + cwd + "/testdata/f4",
				Path: "/opt/funnel/inputs/testdata/f4",
				Type: tes.FileType_DIRECTORY,
			},
			{
				Name:     "c1",
				Path:     "/opt/funnel/inputs/testdata/contents.txt",
				Contents: "test content\n",
			},
		},
		Outputs: []*tes.TaskParameter{
			{
				Name: "stdout-0",
				Url:  "file://" + cwd + "/testdata/stdout-first",
				Path: "/opt/funnel/outputs/stdout-0",
			},
			{
				Name: "o9",
				Url:  "file://" + cwd + "/testdata/o9",
				Path: "/opt/funnel/outputs/sub/o9",
				Type: tes.FileType_DIRECTORY,
			},
		},
		Volumes: []string{"/volone", "/voltwo"},
	}

	err = f.MapTask(task)
	if err != nil {
		t.Fatal(err)
	}

	ei := []*tes.TaskParameter{
		{
			Name: "f1",
			Url:  "file://" + cwd + "/testdata/f1.txt",
			Path: tmp + "/opt/funnel/inputs/testdata/f1.txt",
		},
		{
			Name: "f4",
			Url:  "file://" + cwd + "/testdata/f4",
			Path: tmp + "/opt/funnel/inputs/testdata/f4",
			Type: tes.FileType_DIRECTORY,
		},
	}

	eo := []*tes.TaskParameter{
		{
			Name: "stdout-0",
			Url:  "file://" + cwd + "/testdata/stdout-first",
			Path: tmp + "/opt/funnel/outputs/stdout-0",
		},
		{
			Name: "o9",
			Url:  "file://" + cwd + "/testdata/o9",
			Path: tmp + "/opt/funnel/outputs/sub/o9",
			Type: tes.FileType_DIRECTORY,
		},
	}

	ev := []Volume{
		{
			HostPath:      tmp + "/volone",
			ContainerPath: "/volone",
			Readonly:      false,
		},
		{
			HostPath:      tmp + "/voltwo",
			ContainerPath: "/voltwo",
			Readonly:      false,
		},
		{
			HostPath:      tmp + "/opt/funnel/inputs/testdata/f1.txt",
			ContainerPath: "/opt/funnel/inputs/testdata/f1.txt",
			Readonly:      true,
		},
		{
			HostPath:      tmp + "/opt/funnel/inputs/testdata/f4",
			ContainerPath: "/opt/funnel/inputs/testdata/f4",
			Readonly:      true,
		},
		{
			HostPath:      tmp + "/opt/funnel/inputs/testdata/contents.txt",
			ContainerPath: "/opt/funnel/inputs/testdata/contents.txt",
			Readonly:      true,
		},
		{
			HostPath:      tmp + "/opt/funnel/outputs",
			ContainerPath: "/opt/funnel/outputs",
			Readonly:      false,
		},
	}

	if diff := deep.Equal(f.Inputs, ei); diff != nil {
		t.Log("Expected", fmt.Sprintf("%+v", ei))
		t.Log("Actual", fmt.Sprintf("%+v", f.Inputs))
		for _, d := range diff {
			t.Log("Diff", d)
		}
		t.Fatal("unexpected mapper inputs")
	}

	c, err := ioutil.ReadFile(tmp + "/opt/funnel/inputs/testdata/contents.txt")
	if err != nil {
		t.Fatal(err)
	}
	if string(c) != "test content\n" {
		t.Fatal("unexpected content")
	}

	if diff := deep.Equal(f.Outputs, eo); diff != nil {
		t.Log("Expected", fmt.Sprintf("%+v", eo))
		t.Log("Actual", fmt.Sprintf("%+v", f.Outputs))
		for _, d := range diff {
			t.Log("Diff", d)
		}
		t.Fatal("unexpected mapper outputs")
	}

	if diff := deep.Equal(f.Volumes, ev); diff != nil {
		t.Log("Expected", fmt.Sprintf("%+v", ev))
		t.Log("Actual", fmt.Sprintf("%+v", f.Volumes))
		for _, d := range diff {
			t.Log("Diff", d)
		}
		t.Fatal("unexpected mapper volumes")
	}
}
