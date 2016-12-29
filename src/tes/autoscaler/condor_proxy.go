package autoscaler

import (
	"fmt"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	pbc "tes/autoscaler/proto"
	"tes/scheduler"
)

type CondorProxy struct {
	binPath string
}

// condorProxy provides a TES specific remote API for HTCondor commands.
//
// HTCondor doesn't have a remote API for running commands against the
// master node (e.g. condor_submit). In order to allow remote control
// from TES, a simple proxy service will be running on the condor master
// node, which TES can easily talk to.

func NewCondorProxy(binPath string) *CondorProxy {
	return &CondorProxy{binPath}
}

func (pxy *CondorProxy) Start(port string) {
	// Open a TCP port.
	// Fail hard if it can't be opened.
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		panic("Cannot open port")
	}

	grpcServer := grpc.NewServer()
	pbc.RegisterCondorProxyServer(grpcServer, pxy)

	// Start the gRPC server
	log.Println("TCP+RPC server listening on " + port)
	grpcServer.Serve(lis)
}

func (pxy *CondorProxy) StartWorker(ctx context.Context, req *pbc.StartWorkerRequest) (*pbc.StartWorkerResponse, error) {
	log.Println("Start condor worker")

	conf := fmt.Sprintf(`
		universe = vanilla
		executable = %s
		arguments = -nworkers 1 -master %s
		log = log
		error = err
		output = out
		queue
	`, pxy.binPath, req.SchedAddr)

	log.Printf("Condor submit config: \n%s", conf)

	cmd := exec.Command("condor_submit")
	stdin, _ := cmd.StdinPipe()
	io.WriteString(stdin, conf)
	stdin.Close()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Start()
	return &pbc.StartWorkerResponse{}, nil
}

type CondorProxyClient struct {
	pbc.CondorProxyClient
	conn *grpc.ClientConn
}

// NewClient returns a new Client instance connected to the
// Condor proxy server at a given address (e.g. "localhost:9090")
func NewCondorProxyClient(address string) (*CondorProxyClient, error) {
	// TODO NewRpcConnection shouldn't be in scheduler
	conn, err := scheduler.NewRpcConnection(address)
	if err != nil {
		log.Printf("Error connecting: %s", err)
		return nil, err
	}

	s := pbc.NewCondorProxyClient(conn)
	return &CondorProxyClient{s, conn}, nil
}

// Close the client connection.
func (client *CondorProxyClient) Close() {
	client.conn.Close()
}
