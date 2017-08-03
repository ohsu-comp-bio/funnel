package openstack

import (
	"context"
	"flag"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"github.com/ohsu-comp-bio/funnel/storage"
	"github.com/ohsu-comp-bio/funnel/tests/e2e"
	"io/ioutil"
	"os"
	"testing"
)

var fun *e2e.Funnel
var confPath = flag.String("openstack-e2e-config", "", "OpenStack end-to-end test config file")
var log = logger.New("e2e-openstack")
var conf config.Config

func TestMain(m *testing.M) {
	log.Configure(logger.DebugConfig())
	flag.Parse()

	if *confPath == "" {
		log.Info("Skipping openstack e2e tests, no config")
		os.Exit(0)
	}

	conf = e2e.DefaultConfig()
	if err := config.ParseFile(*confPath, &conf); err != nil {
		panic(err)
	}

	fun = e2e.NewFunnel(conf)
	fun.StartServer()

	os.Exit(m.Run())
}

func TestSwiftStorageTask(t *testing.T) {
	id := fun.Run(`
    --sh 'md5sum $in > $out'
    -i in=swift://buchanan-scratch/funnel
    -o out=swift://buchanan-scratch/funnel-md5
  `)
	task := fun.Wait(id)

	expect := "da385a552397a4ac86ee6444a8f9ae3e  /opt/funnel/inputs/buchanan-scratch/funnel\n"

	if task.State != tes.State_COMPLETE {
		t.Fatal("Unexpected task failure")
	}

	s := storage.Storage{}
	s, serr := s.WithConfig(conf.Worker.Storage)
	if serr != nil {
		t.Fatal("Error configuring storage", serr)
	}

	ctx := context.Background()
	gerr := s.Get(ctx, "swift://buchanan-scratch/funnel-md5", "swift-md5-out", tes.FileType_FILE)
	if gerr != nil {
		t.Fatal("Failed get", gerr.Error())
	}

	b, err := ioutil.ReadFile("swift-md5-out")
	if err != nil {
		t.Fatal("Failed read", err)
	}

	if string(b) != expect {
		t.Fatal("unexpected content", string(b))
	}
}

func TestSwiftDirStorageTask(t *testing.T) {
	id := fun.Run(`
    --sh 'mkdir -p $out/dir; md5sum $in > $out/dir/md5.txt'
    -i in=swift://buchanan-scratch/funnel
    -O out=swift://buchanan-scratch/funnel-md5-dir
  `)
	task := fun.Wait(id)

	expect := "da385a552397a4ac86ee6444a8f9ae3e  /opt/funnel/inputs/buchanan-scratch/funnel\n"

	if task.State != tes.State_COMPLETE {
		t.Fatal("Unexpected task failure")
	}

	s := storage.Storage{}
	s, serr := s.WithConfig(conf.Worker.Storage)
	if serr != nil {
		t.Fatal("Error configuring storage", serr)
	}

	ctx := context.Background()
	gerr := s.Get(ctx, "swift://buchanan-scratch/funnel-md5-dir", "swift-md5-out-dir", tes.FileType_DIRECTORY)
	if gerr != nil {
		t.Fatal("Failed get", gerr.Error())
	}

	b, err := ioutil.ReadFile("swift-md5-out-dir/dir/md5.txt")
	if err != nil {
		t.Fatal("Failed read", err)
	}

	if string(b) != expect {
		t.Fatal("unexpected content", string(b))
	}
}
