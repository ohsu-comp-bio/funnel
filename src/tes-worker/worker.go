package main

import (
	"flag"
	uuid "github.com/nu7hatch/gouuid"
	"google.golang.org/grpc"
	"log"
	"os"
	"path/filepath"
	"strings"
	"tes/server/proto"
	"tes/worker"
	"time"
)

func main() {
	agroServer := flag.String("master", "localhost:9090", "Master Server")
	volumeDirArg := flag.String("volumes", "volumes", "Volume Dir")
	storageDirArg := flag.String("storage", "storage", "Storage Dir")
	fileSystemArg := flag.String("files", "", "Allowed File Paths")
	swiftDirArg := flag.String("swift", "", "Cache Swift items in directory")
	timeoutArg := flag.Int("timeout", -1, "Timeout in seconds")

	nworker := flag.Int("nworkers", 4, "Worker Count")
	flag.Parse()
	volumeDir, _ := filepath.Abs(*volumeDirArg)
	if _, err := os.Stat(volumeDir); os.IsNotExist(err) {
		os.Mkdir(volumeDir, 0700)
	}

	log.Println("Connecting GA4GH Task Server")
	conn, err := grpc.Dial(*agroServer, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	schedClient := ga4gh_task_ref.NewSchedulerClient(conn)

	var fileClient tesTaskEngineWorker.FileSystemAccess

	if *swiftDirArg != "" {
		storageDir, _ := filepath.Abs(*swiftDirArg)
		if _, err := os.Stat(storageDir); os.IsNotExist(err) {
			os.Mkdir(storageDir, 0700)
		}

		fileClient = tesTaskEngineWorker.NewSwiftAccess()
	} else if *fileSystemArg != "" {
		o := []string{}
		for _, i := range strings.Split(*fileSystemArg, ",") {
			p, _ := filepath.Abs(i)
			o = append(o, p)
		}
		fileClient = tesTaskEngineWorker.NewFileAccess(o)
	} else {
		storageDir, _ := filepath.Abs(*storageDirArg)
		if _, err := os.Stat(storageDir); os.IsNotExist(err) {
			os.Mkdir(storageDir, 0700)
		}
		fileClient = tesTaskEngineWorker.NewSharedFS(storageDir)
	}
	fileMapper := tesTaskEngineWorker.NewFileMapper(fileClient, volumeDir)

	u, _ := uuid.NewV4()
	manager, _ := tesTaskEngineWorker.NewLocalManager(*nworker, u.String())
	if *timeoutArg <= 0 {
		manager.Run(schedClient, *fileMapper)
	} else {
		var startCount int32
		lastPing := time.Now().Unix()
		manager.SetStatusCheck(func(status tesTaskEngineWorker.EngineStatus) {
			if status.JobCount > startCount || status.ActiveJobs > 0 {
				startCount = status.JobCount
				lastPing = time.Now().Unix()
			}
		})
		manager.Start(schedClient, *fileMapper)
		for time.Now().Unix()-lastPing < int64(*timeoutArg) {
			time.Sleep(5 * time.Second)
		}

	}
}
