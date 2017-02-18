package gce

import (
	"context"
	"fmt"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/compute/v1"
	"io/ioutil"
	"net/http"
	"tes/config"
)

const startupScriptTpl = `
#!/bin/bash

cat <<CONFYML > $HOME/worker.config.yml
%s
CONFYML

$HOME/task-execution-server/bin/tes-worker -config $HOME/worker.config.yml
`

func (s *scheduler) startWorker(workerID string) {
	ctx := context.Background()
	project := s.conf.Schedulers.GCE.Project
	zone := s.conf.Schedulers.GCE.Zone

	svc, serr := service(ctx, s.conf)
	if serr != nil {
		log.Error("Couldn't create GCE service client", serr)
		return
	}

	tpl, terr := svc.InstanceTemplates.Get(project, s.conf.Schedulers.GCE.Template).Do()
	if terr != nil {
		log.Error("Couldn't retrieve GCE instance template",
			"error", terr,
			"template", s.conf.Schedulers.GCE.Template)
		return
	}

	props := tpl.Properties
	metadata := compute.Metadata{
		Items: append(props.Metadata.Items,
			&compute.MetadataItems{
				Key:   "tes-worker-id",
				Value: &workerID,
			},
			&compute.MetadataItems{
				Key:   "tes-server-address",
				Value: &s.conf.ServerAddress,
			},
		),
	}

	for _, disk := range props.Disks {
		disk.InitializeParams.DiskType = localize(zone, "diskTypes", disk.InitializeParams.DiskType)
	}

	/*
		    w := s.conf.Worker
		    w.ID = workerID
		    w.Timeout = 0
		    w.Storage = s.conf.Storage

				workerConf := worker.Config{
					ID:            workerID,
					ServerAddress: s.conf.ServerAddress,
					Timeout:       -1,
					NumWorkers:    1,
					WorkDir:       "",
				}

				// TODO document that these working dirs need manual cleanup
				//workdir := path.Join(s.conf.WorkDir, "gcp-scheduler", workerID)
				workdir, _ = filepath.Abs(workdir)
				os.MkdirAll(workdir, 0755)
				confPath := path.Join(workdir, "worker.conf.yaml")
				workerConf.ToYamlFile(confPath)
				startupScript := fmt.Sprintf(startupScriptTpl, string(workerConf.ToYaml()))
	*/

	instance := compute.Instance{
		Name:              "tes-worker-" + workerID,
		CanIpForward:      props.CanIpForward,
		Description:       props.Description,
		Disks:             props.Disks,
		MachineType:       localize(zone, "machineTypes", props.MachineType),
		NetworkInterfaces: props.NetworkInterfaces,
		Scheduling:        props.Scheduling,
		ServiceAccounts:   props.ServiceAccounts,
		Tags:              props.Tags,
		Metadata:          &metadata,
	}

	op, ierr := svc.Instances.Insert(project, zone, &instance).Do()
	if ierr != nil {
		log.Error("Couldn't insert GCE VM instance", ierr)
	}
	log.Debug("VM instance", op)
}

// service creates a Google Compute service client
func service(ctx context.Context, conf config.Config) (*compute.Service, error) {
	var client *http.Client
	if conf.Schedulers.GCE.AccountFile != "" {
		// Pull the client configuration (e.g. auth) from a given account file.
		// This is likely downloaded from Google Cloud manually via IAM & Admin > Service accounts.
		bytes, rerr := ioutil.ReadFile(conf.Schedulers.GCE.AccountFile)
		if rerr != nil {
			return nil, rerr
		}

		config, tserr := google.JWTConfigFromJSON(bytes, compute.ComputeScope)
		if tserr != nil {
			return nil, tserr
		}
		client = config.Client(ctx)
	} else {
		// Pull the information (auth and other config) from the environment,
		// which is useful when this code is running in a Google Compute instance.
		client, _ = google.DefaultClient(ctx, compute.ComputeScope)
		// TODO catch error
	}

	svc, cerr := compute.New(client)
	if cerr != nil {
		return nil, cerr
	}
	return svc, nil
}

func localize(zone, resourceType, val string) string {
	return fmt.Sprintf("zones/%s/%s/%s", zone, resourceType, val)
}
