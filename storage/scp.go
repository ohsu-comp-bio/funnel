package storage

import (
	"fmt"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path"
	"strings"
	"time"
)

// SCPClient provides access to remote storage systems
type SCPClient struct {
	Host   string
	Config *ssh.ClientConfig
	Conn   *ssh.Client
	Client *sftp.Client
}

// NewSCPClient returns a SCPClient instance, configured to connect
// to a remote SSH server
func NewSCPClient(host string) *SCPClient {
	host = strings.TrimPrefix(host, "http://")
	host = fmt.Sprintf("%s:%d", host, 22)
	username := os.Getenv("USER")
	var auths []ssh.AuthMethod
	//_ = sshAgent()
	auth := publicKeyFile()
	if auth != nil {
		auths = append(auths, auth)
	}

	config := &ssh.ClientConfig{
		User:            username,
		Auth:            auths,
		Timeout:         10 * time.Second,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// Create a new SCP client
	return &SCPClient{
		Host:   host,
		Config: config,
	}
}

// Connect establishes a connection to the remote SSH server
func (s *SCPClient) Connect() error {
	var err error
	s.Conn, err = ssh.Dial("tcp", s.Host, s.Config)
	if err != nil {
		return err
	}
	s.Client, err = sftp.NewClient(s.Conn)
	if err != nil {
		return fmt.Errorf("unable to start sftp subsytem: %v", err)
	}
	log.Debug("SCP Client Connection Open", "host", s.Host)
	return nil
}

// Close closes the connection to the remote SSH server
func (s *SCPClient) Close() error {
	var err error
	err = s.Conn.Close()
	if err != nil {
		return err
	}
	err = s.Client.Close()
	if err != nil {
		return err
	}
	log.Debug("SCP Client Connection Closed", "host", s.Host)
	return nil
}

// SCPLocalToRemote copies a local file to a remote destination
func (s *SCPClient) SCPLocalToRemote(source string, dest string) error {
	// Connect to the remote server
	err := s.Connect()
	if err != nil {
		return fmt.Errorf("Couldn't establish a connection to the remote server: %v", err)
	}
	// Close connection after the file has been copied
	defer s.Close()

	sf, err := os.Open(source)
	if err != nil {
		return err
	}
	defer sf.Close()

	dstD := path.Dir(dest)
	if _, err := s.Client.Stat(dstD); err != nil {
		err := mkdirAll(dstD, s.Client.Mkdir)
		if err != nil {
			return err
		}
	}
	df, err := s.Client.Create(dest)
	if err != nil {
		return err
	}

	_, err = io.Copy(df, sf)
	cerr := df.Close()
	if err != nil {
		return err
	}
	if cerr != nil {
		return cerr
	}
	return nil
}

// SCPRemoteToLocal copies a remote file to a local destination
func (s *SCPClient) SCPRemoteToLocal(source string, dest string) error {
	// Connect to the remote server
	err := s.Connect()
	if err != nil {
		return fmt.Errorf("Couldn't establish a connection to the remote server: %v ", err)
	}
	// Close connection after the file has been copied
	defer s.Close()

	sf, err := s.Client.Open(source)
	if err != nil {
		return err
	}
	defer sf.Close()

	dstD := path.Dir(dest)
	if _, err := os.Stat(dstD); err != nil {
		mkdirAll(dstD, mkdir)
		if err != nil {
			return err
		}
	}
	df, err := os.Create(dest)
	if err != nil {
		return err
	}

	_, err = io.Copy(df, sf)
	cerr := df.Close()
	if err != nil {
		return err
	}
	if cerr != nil {
		return cerr
	}
	return nil
}

func mkdirAll(p string, mkdir func(string) error) error {
	var current string
	pathParts := strings.Split(p, "/")
	current, pathParts = pathParts[0], pathParts[1:]
	for _, dir := range pathParts {
		current = path.Join(current, dir)
		err := mkdir(current)
		if err != nil {
			return err
		}
	}
	return nil
}

func mkdir(p string) error {
	return os.Mkdir(p, 0777)
}

func publicKeyFile() ssh.AuthMethod {
	home := os.Getenv("HOME")
	file := path.Join(home, ".ssh", "id_rsa")
	buffer, err := ioutil.ReadFile(file)
	if err != nil {
		return nil
	}
	key, err := ssh.ParsePrivateKey(buffer)
	if err != nil {
		return nil
	}
	return ssh.PublicKeys(key)
}

func sshAgent() ssh.AuthMethod {
	sshAgent, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK"))
	if err == nil {
		return ssh.PublicKeysCallback(agent.NewClient(sshAgent).Signers)
	}
	log.Debug("sshAgent", "", err)
	return nil
}
