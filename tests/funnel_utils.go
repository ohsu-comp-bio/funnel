package tests

import (
	"bytes"
	"errors"
	"fmt"
	dockerTypes "github.com/docker/docker/api/types"
	dockerFilters "github.com/docker/docker/api/types/filters"
	docker "github.com/docker/docker/client"
	"github.com/ohsu-comp-bio/funnel/client"
	runlib "github.com/ohsu-comp-bio/funnel/cmd/run"
	servercmd "github.com/ohsu-comp-bio/funnel/cmd/server"
	workercmd "github.com/ohsu-comp-bio/funnel/cmd/worker"
	"github.com/ohsu-comp-bio/funnel/compute/scheduler"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/logger"
	pbs "github.com/ohsu-comp-bio/funnel/proto/scheduler"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"github.com/ohsu-comp-bio/funnel/server"
	"github.com/ohsu-comp-bio/funnel/util"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"text/template"
	"time"
)

var log = logger.NewLogger("e2e", LogConfig())

func init() {
	logger.ConfigureGRPC(1, log)
}

// Funnel provides a test server and RPC/HTTP clients
type Funnel struct {
	// Clients
	RPC    tes.TaskServiceClient
	HTTP   *client.Client
	Docker *docker.Client

	// Config
	Conf       config.Config
	StorageDir string

	// Components
	DB        server.Database
	Server    *server.Server
	Scheduler *scheduler.Scheduler
	SDB       scheduler.Database
	Srv       *servercmd.Server

	// Internal
	startTime string
	rate      time.Duration
	conn      *grpc.ClientConn
}

// NewFunnel creates a new funnel test server with some test
// configuration automatically set: random ports, temp work dir, etc.
func NewFunnel(conf config.Config) *Funnel {
	conf = config.EnsureServerProperties(conf)

	dcli, derr := util.NewDockerClient()
	if derr != nil {
		panic(derr)
	}

	srv, err := servercmd.NewServer(conf, log)
	if err != nil {
		panic(err)
	}

	return &Funnel{
		HTTP:       client.NewClient(conf.Server.HTTPAddress()),
		Docker:     dcli,
		Conf:       conf,
		StorageDir: conf.Worker.Storage.Local.AllowedDirs[0],
		DB:         srv.DB,
		SDB:        srv.SDB,
		Server:     srv.Server,
		Srv:        srv,
		Scheduler:  srv.Scheduler,
		startTime:  fmt.Sprintf("%d", time.Now().Unix()),
		rate:       time.Millisecond * 500,
	}
}

// AddRPCClient configures and connects the RPC client to the server.
func (f *Funnel) AddRPCClient(opts ...grpc.DialOption) error {
	conn, err := NewRPCConn(f.Conf, opts...)
	if err != nil {
		return err
	}
	f.conn = conn
	f.RPC = tes.NewTaskServiceClient(conn)
	return nil
}

// Cleanup cleans up test resources
func (f *Funnel) Cleanup() {
	os.RemoveAll(f.StorageDir)
	os.RemoveAll(f.Conf.Worker.WorkDir)
	f.conn.Close()
}

// StartServer starts the server
func (f *Funnel) StartServer() {
	go f.Srv.Run(context.Background())

	err := f.PollForServerStart()
	if err != nil {
		log.Error("failed to start funnel server", err)
		panic(err)
	}

	err = f.AddRPCClient()
	if err != nil {
		log.Error("failed to add RPC client", err)
		panic(err)
	}

	log.Info("Funnel server started")
}

// PollForServerStart polls the server http address to check if the server is running.
func (f *Funnel) PollForServerStart() error {
	ready := make(chan struct{})
	go func() {
		for {
			_, err := http.Get(f.Conf.Server.HTTPAddress())
			if err == nil {
				close(ready)
				break
			}
			time.Sleep(time.Millisecond * 100)
		}
	}()

	select {
	case <-ready:
		time.Sleep(time.Second * 1)
		return nil
	case <-time.After(time.Second * 30):
		return fmt.Errorf("timeout error - server didn't start within 30 seconds")
	}
}

// WaitForDockerDestroy waits for a "destroy" event
// from docker for the given container ID
//
// TODO probably could use docker.ContainerWait()
// https://godoc.org/github.com/moby/moby/client#Client.ContainerWait
func (f *Funnel) WaitForDockerDestroy(id string) {
	fil := dockerFilters.NewArgs()
	fil.Add("type", "container")
	fil.Add("container", id)
	fil.Add("event", "destroy")

	s, err := f.Docker.Events(context.Background(), dockerTypes.EventsOptions{
		Since:   string(f.startTime),
		Filters: fil,
	})
	for {
		select {
		case e := <-err:
			panic(e)
		case <-s:
			return
		}
	}
}

// Cancel cancels a task by ID
func (f *Funnel) Cancel(id string) error {
	_, err := f.RPC.CancelTask(context.Background(), &tes.CancelTaskRequest{
		Id: id,
	})
	return err
}

// ListView returns a task list (calls ListTasks) with the given view.
func (f *Funnel) ListView(view tes.TaskView) []*tes.Task {
	t, err := f.RPC.ListTasks(context.Background(), &tes.ListTasksRequest{
		View: view,
	})
	if err != nil {
		panic(err)
	}
	return t.Tasks
}

// GetView returns a task (calls GetTask) with the given view.
func (f *Funnel) GetView(id string, view tes.TaskView) *tes.Task {
	t, err := f.RPC.GetTask(context.Background(), &tes.GetTaskRequest{
		Id:   id,
		View: view,
	})
	if err != nil {
		panic(err)
	}
	return t
}

// Get gets a task by ID
func (f *Funnel) Get(id string) *tes.Task {
	return f.GetView(id, tes.TaskView_FULL)
}

// Run uses `funnel run` syntax to create a task.
// Run panics on error.
func (f *Funnel) Run(s string) string {
	id, err := f.RunE(s)
	if err != nil {
		panic(err)
	}
	return id
}

// RunE is like Run(), but returns an error instead of panicing.
func (f *Funnel) RunE(s string) (string, error) {
	// Process the string as a template to allow a few helpers
	tpl := template.Must(template.New("run").Parse(s))
	var by bytes.Buffer
	data := map[string]string{
		"storage": "./" + f.StorageDir,
	}
	if eerr := tpl.Execute(&by, data); eerr != nil {
		return "", eerr
	}
	s = by.String()

	tasks, err := runlib.ParseString(s)
	if err != nil {
		return "", err
	}
	if len(tasks) > 1 {
		return "", fmt.Errorf("Funnel run only handles a single task (no scatter)")
	}
	return f.RunTask(tasks[0])
}

// RunTask calls CreateTask with the given task message and returns the ID.
func (f *Funnel) RunTask(t *tes.Task) (string, error) {
	resp, cerr := f.RPC.CreateTask(context.Background(), t, grpc.FailFast(false))
	if cerr != nil {
		return "", cerr
	}
	return resp.Id, nil
}

// Wait waits for a task to complete.
func (f *Funnel) Wait(id string) *tes.Task {
	for range time.NewTicker(f.rate).C {
		t := f.Get(id)
		if t.State != tes.State_QUEUED && t.State != tes.State_INITIALIZING &&
			t.State != tes.State_RUNNING {
			return t
		}
	}
	return nil
}

// WaitForRunning waits for a task to be in the RUNNING state
func (f *Funnel) WaitForRunning(ids ...string) {
	for _, id := range ids {
		for range time.NewTicker(f.rate).C {
			t := f.Get(id)
			if t.State == tes.State_RUNNING {
				break
			}
		}
	}
}

// WaitForExec waits for a task to reach the given executor index.
// 1 is the first executor.
func (f *Funnel) WaitForExec(id string, i int) {
	for range time.NewTicker(f.rate).C {
		t := f.Get(id)
		if len(t.Logs[0].Logs) >= i {
			return
		}
	}
}

// Tempdir returns a new temporary directory path
func (f *Funnel) Tempdir() string {
	d, _ := ioutil.TempDir(f.StorageDir, "")
	d, _ = filepath.Abs(d)
	return d
}

// WriteFile writes a file to the local (temporary) storage directory.
func (f *Funnel) WriteFile(name string, content string) {
	err := ioutil.WriteFile(f.StorageDir+"/"+name, []byte(content), os.ModePerm)
	if err != nil {
		panic(err)
	}
}

// ReadFile reads a file to the local (temporary) storage directory.
func (f *Funnel) ReadFile(name string) string {
	b, err := ioutil.ReadFile(f.StorageDir + "/" + name)
	if err != nil {
		panic(err)
	}
	return string(b)
}

// StartServerInDocker starts a funnel server in a docker image
func (f *Funnel) StartServerInDocker(imageName string, extraArgs []string) {
	// find the funnel-linux-amd64 binary
	// TODO there must be a better way than this hardcoded path
	funnelBinary, err := filepath.Abs(filepath.Join(
		"../../build/bin/funnel-linux-amd64",
	))
	if err != nil {
		log.Error("Error finding funnel-linux-amd64 binary. Run `make cross-compile`", err)
		panic(err)
	}
	fi, err := os.Stat(funnelBinary)
	if os.IsNotExist(err) || fi.Mode().IsDir() || !strings.Contains(fi.Mode().String(), "x") {
		log.Error("Error finding funnel-linux-amd64 binary. Run `make cross-compile`")
		panic(errors.New(""))
	}

	// write config file
	configPath, _ := filepath.Abs(filepath.Join(f.Conf.Worker.WorkDir, "config.yml"))
	f.Conf.ToYamlFile(configPath)
	os.Chmod(configPath, 0644)

	httpPort, _ := strconv.ParseInt(f.Conf.Server.HTTPPort, 0, 32)
	rpcPort, _ := strconv.ParseInt(f.Conf.Server.RPCPort, 0, 32)

	// detect gid of /var/run/docker.sock
	fi, err = os.Stat("/var/run/docker.sock")
	if err != nil {
		panic(err)
	}
	gid := fi.Sys().(*syscall.Stat_t).Gid

	// setup docker run cmd
	args := []string{
		"run", "-i", "--rm",
		"--group-add", fmt.Sprintf("%d", gid),
		"--name", "funnel-test-server-" + RandomString(6),
		"-p", fmt.Sprintf("%d:%d", httpPort, httpPort),
		"-p", fmt.Sprintf("%d:%d", rpcPort, rpcPort),
		"-v", "/var/run/docker.sock:/var/run/docker.sock",
		"-v", fmt.Sprintf("%s:/bin/funnel", funnelBinary),
		"-v", fmt.Sprintf("%s:%s", configPath, configPath),
	}
	args = append(args, extraArgs...)
	args = append(args, imageName, "funnel", "server", "run", "--config", configPath)

	cmd := exec.Command("docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	log.Info("Running command", "cmd", "docker "+strings.Join(args, " "))

	// start server
	done := make(chan error, 1)
	go func() {
		done <- cmd.Run()
	}()

	ready := make(chan struct{})
	go func() {
		err := f.PollForServerStart()
		if err != nil {
			log.Error("failed to start funnel server", err)
			panic(err)
		}

		err = f.AddRPCClient()
		if err != nil {
			log.Error("failed to add RPC client", err)
			panic(err)
		}

		close(ready)
	}()

	select {
	case err := <-done:
		log.Error("Error starting funnel server in container", err)
		panic(err)
	case <-ready:
		break
	}
	return
}

func (f *Funnel) findTestServerContainers() []string {
	res := []string{}
	containers, err := f.Docker.ContainerList(context.Background(), dockerTypes.ContainerListOptions{})
	if err != nil {
		panic(err)
	}
	for _, c := range containers {
		for _, n := range c.Names {
			if strings.Contains(n, "funnel-test-server-") {
				res = append(res, n)
			}
		}
	}
	return res
}

func (f *Funnel) killTestServerContainers(ids []string) {
	timeout := 10 * time.Second
	for _, n := range ids {
		err := f.Docker.ContainerStop(context.Background(), strings.TrimPrefix(n, "/"), &timeout)
		if err != nil {
			panic(err)
		}
	}
	return
}

// CleanupTestServerContainer stops the docker container running the test funnel server
func (f *Funnel) CleanupTestServerContainer() {
	f.Cleanup()
	s := f.findTestServerContainers()
	f.killTestServerContainers(s)
	return
}

// ListNodes calls db.ListNodes.
func (f *Funnel) ListNodes() []*pbs.Node {
	resp, _ := f.DB.ListNodes(context.Background(), &pbs.ListNodesRequest{})
	return resp.Nodes
}

// AddNode starts an in-memory node routine.
func (f *Funnel) AddNode(conf config.Config) {
	n, err := scheduler.NewNode(conf, logger.NewLogger("e2e-node", conf.Scheduler.Node.Logger), workercmd.NewDefaultWorker)
	if err != nil {
		panic(err)
	}
	go n.Run(context.Background())
	time.Sleep(time.Second * 2)
}

// NewRPCConn returns a new grpc.ClientConn, to make creating TES clients easier.
func NewRPCConn(conf config.Config, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	opts = append(opts,
		grpc.WithInsecure(),
		grpc.WithBlock(),
	)

	conn, err := grpc.DialContext(
		ctx,
		conf.Server.RPCAddress(),
		opts...,
	)
	if err != nil {
		return nil, err
	}

	return conn, nil
}
