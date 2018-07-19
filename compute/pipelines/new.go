package pipelines

import (
	"context"
	"io/ioutil"
	"net/http"

	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/events"
	"github.com/ohsu-comp-bio/funnel/tes"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/genomics/v2alpha1"
)

// Backend provides a compute backend for submitting tasks to
// Google Cloud Genomics Pipelines API:
// https://cloud.google.com/genomics/reference/rest/v1alpha2/pipelines
type Backend struct {
	client   *genomics.Service
	conf     config.Pipelines
	event    events.Writer
	database tes.ReadOnlyServer
}

// NewBackend creates the compute backend and starts a background goroutine
// for monitoring tasks.
func NewBackend(ctx context.Context, conf config.Pipelines, writer events.Writer, database tes.ReadOnlyServer) (*Backend, error) {
	client, err := newClient(ctx, conf)
	if err != nil {
		return nil, err
	}

	svc, err := genomics.New(client)
	if err != nil {
		return nil, err
	}

	b := &Backend{svc, conf, writer, database}
	go b.reconcile(ctx)
	return b, nil
}

func newClient(ctx context.Context, conf config.Pipelines) (*http.Client, error) {
	client := &http.Client{}

	if conf.CredentialsFile != "" {
		// Pull the client configuration (e.g. auth) from a given account file.
		// This is likely downloaded from Google Cloud manually via IAM & Admin > Service accounts.
		bytes, rerr := ioutil.ReadFile(conf.CredentialsFile)
		if rerr != nil {
			return nil, rerr
		}

		config, tserr := google.JWTConfigFromJSON(bytes, genomics.CloudPlatformScope)
		if tserr != nil {
			return nil, tserr
		}
		client = config.Client(ctx)
	} else {
		// Pull the information (auth and other config) from the environment,
		// which is useful when this code is running in a Google Compute instance.
		defClient, err := google.DefaultClient(ctx, genomics.CloudPlatformScope)
		if err == nil {
			client = defClient
		}
	}

	return client, nil
}
