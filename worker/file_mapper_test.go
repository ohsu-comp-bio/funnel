package worker

import (
	"fmt"
	"github.com/go-test/deep"
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"io/ioutil"
	"os"
	"testing"
)

func init() {
	logger.ForceColors()
}

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
			HostPath: tmp + "/volone",
			ContainerPath: "/volone",
			Readonly: false,
		},
		{
			HostPath: tmp + "/voltwo",
			ContainerPath: "/voltwo",
			Readonly: false,
		},
		{
			HostPath: tmp + "/opt/funnel/inputs/testdata/f1.txt",
			ContainerPath: "/opt/funnel/inputs/testdata/f1.txt",
			Readonly: true,
		},
		{
			HostPath: tmp + "/opt/funnel/inputs/testdata/f4",
			ContainerPath: "/opt/funnel/inputs/testdata/f4",
			Readonly: true,
		},
		{
			HostPath: tmp + "/opt/funnel/inputs/testdata/contents.txt",
			ContainerPath: "/opt/funnel/inputs/testdata/contents.txt",
			Readonly: true,
		},
		{
			HostPath: tmp + "/opt/funnel/outputs",
			ContainerPath: "/opt/funnel/outputs",
			Readonly: false,
		},
	}

	if diff := deep.Equal(f.Inputs, ei); diff != nil {
		log.Debug("Expected", fmt.Sprintf("%+v", ei))
		log.Debug("Actual", fmt.Sprintf("%+v", f.Inputs))
		for _, d := range diff {
			log.Debug("Diff", d)
		}
		t.Fatal("unexpected mapper inputs")
	}

	if diff := deep.Equal(f.Outputs, eo); diff != nil {
		t.Fatal("unexpected mapper inputs")
	}

	if diff := deep.Equal(f.Volumes, ev); diff != nil {
		log.Debug("Expected", fmt.Sprintf("%+v", ev))
		log.Debug("Actual", fmt.Sprintf("%+v", f.Volumes))
		for _, d := range diff {
			log.Debug("Diff", d)
		}
		t.Fatal("unexpected mapper volumes")
	}
}
