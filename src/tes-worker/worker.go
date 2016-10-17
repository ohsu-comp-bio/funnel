package main

import (
	"flag"
	uuid "github.com/nu7hatch/gouuid"
	"google.golang.org/grpc"
	"log"
	"os"
	"strings"
	"path/filepath"
	"tes/server/proto"
	"tes/worker"
	"time"
)

func main() {
	agro_server := flag.String("master", "localhost:9090", "Master Server")
	volume_dir_arg := flag.String("volumes", "volumes", "Volume Dir")
	storage_dir_arg := flag.String("storage", "storage", "Storage Dir")
	file_system_arg := flag.String("files", "", "Allowed File Paths")
	swift_dir_arg := flag.String("swift", "", "Cache Swift items in directory")
	timeout_arg := flag.Int("timeout", -1, "Timeout in seconds")

	nworker := flag.Int("nworkers", 4, "Worker Count")
	flag.Parse()
	volume_dir, _ := filepath.Abs(*volume_dir_arg)
	if _, err := os.Stat(volume_dir); os.IsNotExist(err) {
		os.Mkdir(volume_dir, 0700)
	}

	log.Println("Connecting GA4GH Task Server")
	conn, err := grpc.Dial(*agro_server, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	sched_client := ga4gh_task_ref.NewSchedulerClient(conn)

	var file_client tes_taskengine_worker.FileSystemAccess = nil

	if *swift_dir_arg != "" {
		storage_dir, _ := filepath.Abs(*swift_dir_arg)
		if _, err := os.Stat(storage_dir); os.IsNotExist(err) {
			os.Mkdir(storage_dir, 0700)
		}

		file_client = tes_taskengine_worker.NewSwiftAccess()
	} else if *file_system_arg != "" {
		o := []string{}
		for _, i := range strings.Split(*file_system_arg, ",") {
			p, _ := filepath.Abs(i)
			o = append(o, p)
		}
		file_client = tes_taskengine_worker.NewFileAccess(o)
	} else {
		storage_dir, _ := filepath.Abs(*storage_dir_arg)
		if _, err := os.Stat(storage_dir); os.IsNotExist(err) {
			os.Mkdir(storage_dir, 0700)
		}
		file_client = tes_taskengine_worker.NewSharedFS(storage_dir)
	}
	file_mapper := tes_taskengine_worker.NewFileMapper(&sched_client, file_client, volume_dir)

	u, _ := uuid.NewV4()
	manager, _ := tes_taskengine_worker.NewLocalManager(*nworker, u.String())
	if *timeout_arg <= 0 {
		manager.Run(sched_client, *file_mapper)
	} else {
		var start_count int32 = 0
		last_ping := time.Now().Unix()
		manager.SetStatusCheck(func(status tes_taskengine_worker.EngineStatus) {
			if status.JobCount > start_count || status.ActiveJobs > 0 {
				start_count = status.JobCount
				last_ping = time.Now().Unix()
			}
		})
		manager.Start(sched_client, *file_mapper)
		for time.Now().Unix()-last_ping < int64(*timeout_arg) {
			time.Sleep(5 * time.Second)
		}

	}
}
