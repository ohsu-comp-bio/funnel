package openstack

import (
	"github.com/rackspace/gophercloud"
	"github.com/rackspace/gophercloud/openstack"
	"github.com/rackspace/gophercloud/openstack/compute/v2/extensions/keypairs"
	"github.com/rackspace/gophercloud/openstack/compute/v2/servers"
	pbr "tes/server/proto"
)

const startScript = `
#!/bin/sh
sudo systemctl start tes.service
`

// StartWorker calls out to OpenStack APIs to start a new worker instance.
func (s *scheduler) StartWorker(w *pbr.Worker) error {

	// TODO move to client wrapper
	authOpts, aerr := openstack.AuthOptionsFromEnv()
	if aerr != nil {
		log.Error("Auth options failed", aerr)
		return aerr
	}

	provider, perr := openstack.AuthenticatedClient(authOpts)
	if perr != nil {
		log.Error("Provider failed", perr)
		return perr
	}

	client, cerr := openstack.NewComputeV2(provider,
		gophercloud.EndpointOpts{Type: "compute", Name: "nova"})

	if cerr != nil {
		log.Error("Provider failed", cerr)
		return cerr
	}

	conf := s.conf.Worker
	conf.ID = w.Id
	conf.ServerAddress = s.conf.ServerAddress
	conf.Storage = s.conf.Storage

	osconf := s.conf.Schedulers.OpenStack

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
					Contents: []byte(conf.ToYaml()),
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
		return serr
	}
	return nil
}

// ShouldStartWorker tells the scaler loop which workers
// belong to this scheduler backend, basically.
func (s *scheduler) ShouldStartWorker(w *pbr.Worker) bool {
	// Only start works that are uninitialized and have a gce template.
	tpl, ok := w.Metadata["openstack"]
	return ok && tpl != "" && w.State == pbr.WorkerState_Uninitialized
}
