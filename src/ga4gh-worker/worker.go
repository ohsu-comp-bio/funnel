
package main

import (
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
	workdir_arg := flag.String("workdir", "/tmp/ga4gh_task_work", "Workdir")
	timeout_arg := flag.Int("timeout", -1, "Timeout in seconds")

	nworker := flag.Int("nworkers", 4, "Worker Count")
	flag.Parse()
	work_dir, _ := filepath.Abs(*workdir_arg)
	log.Println("Connecting GA4GH Task Server")
	conn, err := grpc.Dial(*agro_server, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	sched_client := ga4gh_task_ref.NewSchedulerClient(conn)

	file_client := ga4gh_taskengine.NewSharedFS(&sched_client, *workdir_arg)

	u, _ := uuid.NewV4()
	manager, _ := ga4gh_taskengine.NewLocalManager(*nworker, work_dir, u.String())
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