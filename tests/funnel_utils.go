package tests

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"text/template"
	"time"

	dockerTypes "github.com/docker/docker/api/types"
	dockerFilters "github.com/docker/docker/api/types/filters"
	docker "github.com/docker/docker/client"
	runlib "github.com/ohsu-comp-bio/funnel/cmd/run"
	servercmd "github.com/ohsu-comp-bio/funnel/cmd/server"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/config/testconfig"
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/ohsu-comp-bio/funnel/server"
	"github.com/ohsu-comp-bio/funnel/tes"
	"github.com/ohsu-comp-bio/funnel/util/dockerutil"
	"github.com/ohsu-comp-bio/funnel/util/rpc"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var log = logger.NewLogger("e2e", testconfig.LogConfig())

func init() {
	logger.SetGRPCLogger(log)
	// nanoseconds are important because the tests run faster than a millisecond
	// which can cause port conflicts
	rand.Seed(time.Now().UTC().UnixNano())
}

// Funnel provides a test server and RPC/HTTP clients
type Funnel struct {
	// Clients
	RPC    tes.TaskServiceClient
	HTTP   *tes.Client
	Docker *docker.Client

	// Config
	Conf       config.Config
	StorageDir string

	// Components
	Server *server.Server

	// Internal
	startTime string
	rate      time.Duration
	conn      *grpc.ClientConn
}

// NewFunnel creates a new funnel test server with some test
// configuration automatically set: random ports, temp work dir, etc.
func NewFunnel(conf config.Config) *Funnel {
	cli, err := tes.NewClient(conf.Server.HTTPAddress())
	if err != nil {
		panic(err)
	}

	dcli, derr := dockerutil.NewDockerClient()
	if derr != nil {
		panic(derr)
	}

	srv, err := servercmd.NewServer(context.Background(), conf, log)
	if err != nil {
		panic(err)
	}

	return &Funnel{
		HTTP:       cli,
		Docker:     dcli,
		Conf:       conf,
		StorageDir: conf.LocalStorage.AllowedDirs[0],
		Server:     srv,
		startTime:  fmt.Sprintf("%d", time.Now().Unix()),
		rate:       time.Millisecond * 500,
	}
}

// addRPCClient configures and connects the RPC client to the server.
func (f *Funnel) addRPCClient() {
	conn, err := rpc.Dial(context.Background(), f.Conf.Server)
	if err != nil {
		panic(err)
	}
	f.RPC = tes.NewTaskServiceClient(conn)
	f.conn = conn
}

// Cleanup cleans up test resources
func (f *Funnel) Cleanup() {
	os.RemoveAll(f.StorageDir)
	os.RemoveAll(f.Conf.Worker.WorkDir)
	f.conn.Close()
}

// StartServer starts the server
func (f *Funnel) StartServer() {
	go func() {
		f.Server.Run(context.Background())
	}()

	err := f.PollForServerStart()
	if err != nil {
		log.Error("failed to start funnel server", err)
		panic(err)
	}

	f.addRPCClient()
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

// WaitForInitializing waits for a task to be in the Initializing state
func (f *Funnel) WaitForInitializing(ids ...string) {
	for _, id := range ids {
		for range time.NewTicker(f.rate).C {
			t := f.Get(id)
			if t.State == tes.State_INITIALIZING {
				break
			}
		}
	}
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
func (f *Funnel) StartServerInDocker(containerName, imageName string, extraArgs []string) {
	var funnelBinary string
	var err error
	gopath := os.Getenv("GOPATH")

	if runtime.GOOS == "linux" {
		funnelBinary, err = filepath.Abs(filepath.Join(
			gopath, "bin/", "funnel",
		))
	} else {
		funnelBinary, err = filepath.Abs(filepath.Join(
			gopath, "src/github.com/ohsu-comp-bio/funnel/build/bin/", "funnel-linux-amd64",
		))
	}
	if err != nil {
		log.Error("Error finding funnel binary", err)
		panic(err)
	}

	// write config file
	configPath, _ := filepath.Abs(filepath.Join(f.Conf.Worker.WorkDir, "config.yml"))
	config.ToYamlFile(f.Conf, configPath)
	os.Chmod(configPath, 0644)

	httpPort, _ := strconv.ParseInt(f.Conf.Server.HTTPPort, 0, 32)
	rpcPort, _ := strconv.ParseInt(f.Conf.Server.RPCPort, 0, 32)

	// detect gid of /var/run/docker.sock
	fi, err := os.Stat("/var/run/docker.sock")
	if err != nil {
		panic(err)
	}
	gid := fi.Sys().(*syscall.Stat_t).Gid

	// setup docker run cmd
	args := []string{
		"run", "-i", "--rm",
		"--group-add", fmt.Sprintf("%d", gid),
		"--name", containerName,
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

		f.addRPCClient()
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
func (f *Funnel) CleanupTestServerContainer(containerName string) {
	f.Cleanup()
	f.killTestServerContainers([]string{containerName})
	return
}

// HelloWorld is a simple, valid task that is easy to reuse in tests.
func HelloWorld() *tes.Task {
	return &tes.Task{
		Id: tes.GenerateID(),
		Executors: []*tes.Executor{
			{
				Image:   "alpine",
				Command: []string{"echo", "hello world"},
			},
		},
	}
}

// SetLogOutput provides a method for connecting funnel logso the the test logger
func SetLogOutput(l *logger.Logger, t *testing.T) {
	l.SetOutput(TestingWriter(t))
}

// TestingWriter returns an io.Writer that writes each line via t.Log
func TestingWriter(t *testing.T) io.Writer {
	reader, writer := io.Pipe()
	scanner := bufio.NewScanner(reader)
	go func() {
		for scanner.Scan() {
			// Carriage return removes testing's file:line number and indent.
			// In this case, the file and line will always be "utils.go:62".
			// Go 1.9 introduced t.Helper() to fix this, but something about
			// this function being in a goroutine seems to break that.
			// Carriage return is the hack for now.
			t.Log("\r" + scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			t.Error("testing writer scanner error", err)
		}
	}()
	return writer
}

// RandomString generates a random string of length n
func RandomString(n int) string {
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyz0123456789")
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
