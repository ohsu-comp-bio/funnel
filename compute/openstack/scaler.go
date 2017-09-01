package openstack

import (
	pbs "github.com/ohsu-comp-bio/funnel/proto/scheduler"
	"github.com/rackspace/gophercloud"
	"github.com/rackspace/gophercloud/openstack"
	"github.com/rackspace/gophercloud/openstack/compute/v2/extensions/keypairs"
	"github.com/rackspace/gophercloud/openstack/compute/v2/servers"
)

const startScript = `
#!/bin/sh
sudo systemctl start tes.service
`

// StartNode calls out to OpenStack APIs to start a new node instance.
func (s *Backend) StartNode(w *pbs.Node) error {

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

	conf := s.conf
	conf.Scheduler.Node.ID = w.Id
	osconf := conf.Backends.OpenStack

	_, serr := servers.Create(client, keypairs.CreateOptsExt{
		CreateOptsBuilder: servers.CreateOpts{
			Name:       osconf.Server.Name,
			FlavorName: osconf.Server.FlavorName,
			ImageName:  osconf.Server.ImageName,
			Networks:   osconf.Server.Networks,
			// Personality defines files that will be copied to the VM instance on boot.
			// We use this to upload Funnel node config.
			Personality: []*servers.File{
				{
					Path:     osconf.ConfigPath,
					Contents: []byte(conf.ToYaml()),
				},
			},
			// Write a simple bash script that starts the Funnel service.
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

// ShouldStartNode tells the scaler loop which nodes
// belong to this scheduler backend, basically.
func (s *Backend) ShouldStartNode(n *pbs.Node) bool {
	// Only start works that are uninitialized and have a template.
	tpl, ok := n.Metadata["openstack"]
	return ok && tpl != "" && n.State == pbs.NodeState_UNINITIALIZED
}
