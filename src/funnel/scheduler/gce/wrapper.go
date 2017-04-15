package gce

import (
	"context"
	"funnel/config"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/compute/v1"
	"io/ioutil"
	"net/http"
)

// Wrapper represents a simpler version of the Google Cloud Compute Service
// interface provided by the google library. The GCE client API is very complex
// and hard to test against, so this wrapper simplifies things down to only what
// funnel needs.
type Wrapper interface {
	InsertInstance(project, zone string, instance *compute.Instance) (*compute.Operation, error)
	ListMachineTypes(project, zone string) (*compute.MachineTypeList, error)
	ListInstanceTemplates(project string) (*compute.InstanceTemplateList, error)
}

func newWrapper(ctx context.Context, conf config.Config) (Wrapper, error) {
	var client *http.Client
	if conf.Backends.GCE.AccountFile != "" {
		// Pull the client configuration (e.g. auth) from a given account file.
		// This is likely downloaded from Google Cloud manually via IAM & Admin > Service accounts.
		bytes, rerr := ioutil.ReadFile(conf.Backends.GCE.AccountFile)
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

	return &wrapper{svc}, nil
}

type wrapper struct {
	svc *compute.Service
}

func (w *wrapper) InsertInstance(project, zone string, instance *compute.Instance) (*compute.Operation, error) {
	return w.svc.Instances.Insert(project, zone, instance).Do()
}

func (w *wrapper) ListMachineTypes(project, zone string) (*compute.MachineTypeList, error) {
	return w.svc.MachineTypes.List(project, zone).Do()
}

func (w *wrapper) ListInstanceTemplates(project string) (*compute.InstanceTemplateList, error) {
	return w.svc.InstanceTemplates.List(project).Do()
}
