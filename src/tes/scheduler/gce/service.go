package gce

import (
	"context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/compute/v1"
	"io/ioutil"
	"net/http"
	"tes/config"
)

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
