package openstack

import (
	"github.com/ghodss/yaml"
	"github.com/rackspace/gophercloud"
	"github.com/rackspace/gophercloud/openstack"
	"github.com/rackspace/gophercloud/openstack/compute/v2/extensions/keypairs"
	"github.com/rackspace/gophercloud/openstack/compute/v2/servers"
	"log"
)

const startScript = `
#!/bin/sh
sudo systemctl start tes.service
`

type Config struct {
	MasterAddr string
	KeyPair    string
	Server     servers.CreateOpts
}

type workerconfig struct {
	MasterAddr string
	ID         string
}

func start(workerID string, config Config) {
	authOpts, aerr := openstack.AuthOptionsFromEnv()
	if aerr != nil {
		log.Printf("Auth options failed")
		log.Println(aerr)
		return
	}

	provider, perr := openstack.AuthenticatedClient(authOpts)
	if perr != nil {
		log.Printf("Provider failed")
		log.Println(perr)
		return
	}

	client, cerr := openstack.NewComputeV2(provider,
		gophercloud.EndpointOpts{Type: "compute", Name: "nova"})

	if cerr != nil {
		log.Printf("Provider failed")
		log.Println(cerr)
		return
	}

	// Write the worker config YAML file, which gets uploaded to the VM.
	wc := workerconfig{
		MasterAddr: config.MasterAddr,
		ID:         workerID,
	}
	wcyml, _ := yaml.Marshal(wc)
	tesConfig := []byte(wcyml)
	// Write a simple bash script that starts the TES service.
	// This will be run when the VM instance boots.
	userData := []byte(startScript)

	_, serr := servers.Create(client, keypairs.CreateOptsExt{
		servers.CreateOpts{
			Name:       config.Server.Name,
			FlavorName: config.Server.FlavorName,
			ImageName:  config.Server.ImageName,
			Networks:   config.Server.Networks,
			// Personality defines files that will be copied to the VM instance on boot.
			// We use this to upload TES worker config.
			Personality: []*servers.File{
				&servers.File{
					// TODO the worker should probably have a list of standard places it
					//      looks for config files, and this should be written to one
					//      of those standard paths.
					Path:     "/home/ubuntu/tes.config.yaml",
					Contents: tesConfig,
				},
			},
			UserData: userData,
		},
		config.KeyPair,
	}).Extract()

	if serr != nil {
		log.Printf("Error creating server")
		log.Println(serr)
	}
}
