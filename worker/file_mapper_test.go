package worker

import (
	"fmt"
	"os"
	"testing"

	"github.com/go-test/deep"
	"github.com/ohsu-comp-bio/funnel/tes"
)

func TestMapTask(t *testing.T) {
	tmp, err := os.MkdirTemp("", "funnel-test-mapper")
	if err != nil {
		t.Fatal(err)
	}
	f := FileMapper{
		WorkDir: tmp,
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	task := &tes.Task{
		Inputs: []*tes.Input{
			{
				Name: "f1",
				Url:  "file://" + cwd + "/testdata/f1.txt",
				Path: "/inputs/testdata/f1.txt",
			},
			{
				Name: "f4",
				Url:  "file://" + cwd + "/testdata/f4",
				Path: "/inputs/testdata/f4",
				Type: tes.FileType_DIRECTORY,
			},
			{
				Name:    "c1",
				Path:    "/inputs/testdata/contents.txt",
				Content: "test content\n",
			},
		},
		Outputs: []*tes.Output{
			{
				Name: "stdout-0",
				Url:  "file://" + cwd + "/testdata/stdout-first",
				Path: "/outputs/stdout-0",
			},
			{
				Name: "o9",
				Url:  "file://" + cwd + "/testdata/o9",
				Path: "/outputs/sub/o9",
				Type: tes.FileType_DIRECTORY,
			},
		},
		Volumes: []string{"/volone", "/voltwo"},
	}

	err = f.MapTask(task)
	if err != nil {
		t.Fatal(err)
	}

	ei := []*tes.Input{
		{
			Name: "f1",
			Url:  "file://" + cwd + "/testdata/f1.txt",
			Path: tmp + "/inputs/testdata/f1.txt",
		},
		{
			Name: "f4",
			Url:  "file://" + cwd + "/testdata/f4",
			Path: tmp + "/inputs/testdata/f4",
			Type: tes.FileType_DIRECTORY,
		},
	}

	eo := []*tes.Output{
		{
			Name: "stdout-0",
			Url:  "file://" + cwd + "/testdata/stdout-first",
			Path: tmp + "/outputs/stdout-0",
		},
		{
			Name: "o9",
			Url:  "file://" + cwd + "/testdata/o9",
			Path: tmp + "/outputs/sub/o9",
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
			HostPath:      tmp + "/tmp",
			ContainerPath: "/tmp",
			Readonly:      false,
		},
		{
			HostPath:      tmp + "/inputs/testdata/f1.txt",
			ContainerPath: "/inputs/testdata/f1.txt",
			Readonly:      true,
		},
		{
			HostPath:      tmp + "/inputs/testdata/f4",
			ContainerPath: "/inputs/testdata/f4",
			Readonly:      true,
		},
		{
			HostPath:      tmp + "/inputs/testdata/contents.txt",
			ContainerPath: "/inputs/testdata/contents.txt",
			Readonly:      true,
		},
		{
			HostPath:      tmp + "/outputs",
			ContainerPath: "/outputs",
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

	c, err := os.ReadFile(tmp + "/inputs/testdata/contents.txt")
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

	if f.ContainerPath(f.Outputs[0].Path) != task.Outputs[0].Path {
		t.Log("Expected", task.Outputs[0].Path)
		t.Log("Actual", f.ContainerPath(f.Outputs[0].Path))
		t.Fatal("path unmapping failed")
	}
}
