package openstack

import (
	"github.com/ghodss/yaml"
	"github.com/rackspace/gophercloud"
	"github.com/rackspace/gophercloud/openstack"
	"github.com/rackspace/gophercloud/openstack/compute/v2/extensions/keypairs"
	"github.com/rackspace/gophercloud/openstack/compute/v2/servers"
)

const startScript = `
#!/bin/sh
sudo systemctl start tes.service
`

func (s *scheduler) start(workerID string) {
	authOpts, aerr := openstack.AuthOptionsFromEnv()
	if aerr != nil {
		log.Error("Auth options failed", aerr)
		return
	}

	provider, perr := openstack.AuthenticatedClient(authOpts)
	if perr != nil {
		log.Error("Provider failed", perr)
		return
	}

	client, cerr := openstack.NewComputeV2(provider,
		gophercloud.EndpointOpts{Type: "compute", Name: "nova"})

	if cerr != nil {
		log.Error("Provider failed", cerr)
		return
	}

	// Write the worker config YAML file, which gets uploaded to the VM.
	workerConf := s.conf.Worker
	workerConf.ID = workerID
	workerConf.ServerAddress = s.conf.ServerConfig.ServerAddress
	workerConf.Storage = s.conf.ServerConfig.Storage
	workerConfYaml, _ := yaml.Marshal(workerConf)

	osconf := s.conf.Schedulers.Openstack
	_, serr := servers.Create(client, keypairs.CreateOptsExt{
		CreateOptsBuilder: servers.CreateOpts{
			Name:       osconf.Server.Name,
			FlavorName: osconf.Server.FlavorName,
			ImageName:  osconf.Server.ImageName,
			Networks:   osconf.Server.Networks,
			// Personality defines files that will be copied to the VM instance on boot.
			// We use this to upload TES worker config.
			Personality: []*servers.File{
				{
					Path:     osconf.ConfigPath,
					Contents: []byte(workerConfYaml),
				},
			},
			// Write a simple bash script that starts the TES service.
			// This will be run when the VM instance boots.
			UserData: []byte(startScript),
		},
		KeyName: osconf.KeyPair,
	}).Extract()

	if serr != nil {
		log.Error("Error creating server", serr)
	}
}
