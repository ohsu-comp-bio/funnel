
package main

import (
	"os"
	"log"
	"flag"
	"ga4gh-engine"
	"google.golang.org/grpc"
	"path/filepath"
	uuid "github.com/nu7hatch/gouuid"
	"time"
	"ga4gh-server/proto"
)


func main() {
	agro_server := flag.String("master", "localhost:9090", "Master Server")
	disk_dir_arg := flag.String("disks", "disks", "Disk Dir")
	storage_dir_arg := flag.String("storage", "storage", "Storage Dir")
	timeout_arg := flag.Int("timeout", -1, "Timeout in seconds")

	nworker := flag.Int("nworkers", 4, "Worker Count")
	flag.Parse()
	storage_dir, _ := filepath.Abs(*storage_dir_arg)
	if _, err := os.Stat(storage_dir); os.IsNotExist(err) {
		os.Mkdir(storage_dir, 0700)
	}
	disk_dir, _ := filepath.Abs(*disk_dir_arg)
	if _, err := os.Stat(disk_dir); os.IsNotExist(err) {
		os.Mkdir(disk_dir, 0700)
	}
	log.Println("Connecting GA4GH Task Server")
	conn, err := grpc.Dial(*agro_server, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	sched_client := ga4gh_task_ref.NewSchedulerClient(conn)

	file_client := ga4gh_taskengine.NewSharedFS(&sched_client, storage_dir, disk_dir )

	u, _ := uuid.NewV4()
	manager, _ := ga4gh_taskengine.NewLocalManager(*nworker, u.String())
	if *timeout_arg <= 0 {
		manager.Run(sched_client, file_client)
	} else {
		var start_count int32 = 0
		last_ping := time.Now().Unix()
		manager.SetStatusCheck( func(status ga4gh_taskengine.EngineStatus) {
			if status.JobCount > start_count || status.ActiveJobs > 0 {
				start_count = status.JobCount
				last_ping = time.Now().Unix()
			}
		} )
		manager.Start(sched_client, file_client)
		for time.Now().Unix() - last_ping < int64(*timeout_arg) {
			time.Sleep(5 * time.Second)
		}

	}
}