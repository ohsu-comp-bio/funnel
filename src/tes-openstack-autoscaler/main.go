package main

import (
	"flag"
  "fmt"
  "log"
  "github.com/rackspace/gophercloud"
  "github.com/rackspace/gophercloud/openstack"
  "github.com/rackspace/gophercloud/openstack/compute/v2/servers"
  "github.com/rackspace/gophercloud/openstack/compute/v2/extensions/keypairs"
  "tes"
)

func main() {
  config := Config{}
	var configArg string
	flag.StringVar(&configArg, "config", "", "Config File")
	flag.Parse()

	tes.LoadConfigOrExit(configArg, &config)
  start(config)
}

type Config struct {
  MasterAddr string
  KeyPair string
  Server servers.CreateOpts
}

func start(config Config) {
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

  // TODO should use yaml marshal lib
  tesConfig := []byte(fmt.Sprintf("MasterAddr: %s\n", config.MasterAddr))
  // Write a simple bash script that starts the TES service.
  // This will be run when the VM instance boots.
  userData := []byte("#!/bin/sh\nsudo systemctl start tes.service")

  _, serr := servers.Create(client, keypairs.CreateOptsExt{
    servers.CreateOpts{
      Name: config.Server.Name,
      FlavorName: config.Server.FlavorName,
      ImageName: config.Server.ImageName,
      Networks: config.Server.Networks,
      // Personality defines files that will be copied to the VM instance on boot.
      // We use this to upload TES worker config.
      Personality: []*servers.File{
        &servers.File{
          // TODO the worker should probably have a list of standard places it
          //      looks for config files, and this should be written to one
          //      of those standard paths.
          Path: "/home/ubuntu/tes.config.yaml",
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
