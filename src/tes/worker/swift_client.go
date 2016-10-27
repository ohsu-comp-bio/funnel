package tesTaskEngineWorker

import (
	"fmt"
	"github.com/rackspace/gophercloud"
	"github.com/rackspace/gophercloud/openstack"
	"io"
	"log"
	"os"
	"strings"
	//"github.com/rackspace/gophercloud/openstack/objectstorage/v1/containers"
	"github.com/rackspace/gophercloud/openstack/objectstorage/v1/objects"
	//"github.com/rackspace/gophercloud/pagination"
)

// SwiftProtocol documentation
// TODO: documentation
var SwiftProtocol = "swift://"

// SwiftAccess documentation
// TODO: documentation
type SwiftAccess struct {
	client *gophercloud.ServiceClient
}

// NewSwiftAccess documentation
// TODO: documentation
func NewSwiftAccess() *SwiftAccess {

	opts, err := openstack.AuthOptionsFromEnv()
	provider, err := openstack.AuthenticatedClient(opts)

	if err != nil {
		panic("Authentication Error")
	}

	swiftClient, err := openstack.NewObjectStorageV1(provider, gophercloud.EndpointOpts{})
	if err != nil {
		panic("Storage Connection Error")
	}

	return &SwiftAccess{client: swiftClient}

}

// Get documentation
// TODO: documentation
func (swiftAccess *SwiftAccess) Get(storage string, hostPath string, class string) error {
	log.Printf("Starting download of %s", storage)
	storage = strings.TrimPrefix(storage, SwiftProtocol)
	storageSplit := strings.SplitN(storage, "/", 2)

	if class == "File" {
		// Download everything into a DownloadResult struct
		opts := objects.DownloadOpts{}
		res := objects.Download(swiftAccess.client, storageSplit[0], storageSplit[1], opts)

		file, err := os.Create(hostPath)
		if err != nil {
			return err
		}
		buffer := make([]byte, 10240)
		totalLen := 0
		for {
			len, err := res.Body.Read(buffer)
			if err == io.EOF {
				break
			} else if err != nil {
				return fmt.Errorf("Error reading file")
			}
			totalLen += len
			file.Write(buffer[:len])
		}
		file.Close()
		res.Body.Close()
		log.Printf("Downloaded %d bytes", totalLen)
		return nil
	} else if class == "Directory" {
		return fmt.Errorf("SWIFT directories not yet supported")
	}
	return fmt.Errorf("Unknown element type: %s", class)
}

// Put documentation
// TODO: documentation
func (swiftAccess *SwiftAccess) Put(storage string, hostPath string, class string) error {
	log.Printf("Starting upload of %s", storage)
	content, err := os.Open(hostPath)
	if err != nil {
		return err
	}

	storage = strings.TrimPrefix(storage, SwiftProtocol)
	storageSplit := strings.SplitN(storage, "/", 2)

	if class == "File" {
		// Now execute the upload
		opts := objects.CreateOpts{}
		res := objects.Create(swiftAccess.client, storageSplit[0], storageSplit[1], content, opts)
		_, err = res.ExtractHeader()
		content.Close()
		return err
	} else if class == "Directory" {
		return fmt.Errorf("SWIFT directories not yet supported")
	}
	return fmt.Errorf("Unknown element type: %s", class)
}
