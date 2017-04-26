package gce

import (
	"encoding/json"
	"fmt"
	"github.com/imdario/mergo"
	"github.com/ohsu-comp-bio/funnel/cmd/server"
	"github.com/ohsu-comp-bio/funnel/cmd/worker"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/ohsu-comp-bio/funnel/logger/logutils"
	"github.com/spf13/cobra"
	"io/ioutil"
	"net/http"
	"strings"
)

var log = logger.New("gce cmd")
var metaURL = "http://metadata.google.internal"

// Cmd represents the 'funnel gce" CLI command set.
var Cmd = &cobra.Command{
	Use: "gce",
}

func init() {
	Cmd.AddCommand(startCmd)
}

var startCmd = &cobra.Command{
	Use: "start",
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error
		conf := config.DefaultConfig()
		conf, err = WithGCEConfig(conf, metaURL)
		if err != nil {
			return err
		}
		logutils.Configure(conf)
		if conf.Worker.ID != "" {
			return worker.Run(conf)
		}
		return server.Run(conf)
	},
}

// WithGCEConfig loads config from a GCE VM environment, in particular
// loading metadata from GCE's instance metadata.
// https://cloud.google.com/compute/docs/storing-retrieving-metadata
func WithGCEConfig(conf config.Config, metaURL string) (config.Config, error) {
	log.Info("Discovering GCE environment")

	// Check that this is a GCE VM environment.
	// If not, fail.
	meta, err := getMetadata(metaURL)
	if err != nil {
		log.Error("Error getting GCE metadata", err)
		return conf, fmt.Errorf("can't find GCE metadata. This command requires a GCE environment")
	}

	log.Info("Loaded GCE metadata")
	log.Debug("GCE metadata", meta)

	conf.Scheduler = "gce"
	defaultHostName := conf.HostName

	// Load full config doc from metadata
	if meta.Instance.Attributes.FunnelConfig != "" {
		mconf := config.Config{}
		var err error
		b := []byte(meta.Instance.Attributes.FunnelConfig)
		err = config.Parse(b, &mconf)
		if err != nil {
			return conf, err
		}
		err = mergo.MergeWithOverwrite(&conf, mconf)
		if err != nil {
			return conf, err
		}
	}

	// Is this a worker node? If so, inherit the worker ID from the GCE instance name.
	if meta.Instance.Attributes.FunnelWorkerServerAddress != "" {
		if conf.Worker.ID == "" {
			conf.Worker.ID = meta.Instance.Name
		}
		conf.Worker.ServerAddress = meta.Instance.Attributes.FunnelWorkerServerAddress
	}
	if meta.Project.ProjectID != "" {
		conf.Backends.GCE.Project = meta.Project.ProjectID
	}
	// TODO need to parse zone?
	if meta.Instance.Zone != "" {
		zone := meta.Instance.Zone
		// Parse zone out of metadata format
		// e.g. "projects/1234/zones/us-west1-b" => "us-west1-b"
		idx := strings.LastIndex(zone, "/")
		if idx != -1 {
			zone = zone[idx+1:]
		}
		conf.Backends.GCE.Zone = zone
	}

	conf.Worker.Metadata["gce"] = "yes"

	// If the configuration contains a worker ID, assume that a worker
	// process should be started (instead of a server).
	if conf.Worker.ID != "" {
		if conf.Worker.ServerAddress == "" {
			log.Error("Empty server address while starting worker")
			return conf, fmt.Errorf("Empty server address while starting worker")
		}
		conf.Worker.Storage = append(conf.Worker.Storage, &config.StorageConfig{
			GS: config.GSStorage{
				FromEnv: true,
			},
		})
	}

	// Auto detect the server's host name when it's not already set.
	// This makes server deployment and configuration a bit easier.
	// TODO will this work across zones?
	if conf.HostName == defaultHostName && meta.Instance.Hostname != "" {
		conf.HostName = meta.Instance.Hostname
	}

	return conf, nil
}

type metadata struct {
	Instance struct {
		Name       string
		Hostname   string
		Zone       string
		Attributes struct {
			FunnelConfig              string `json:"funnel-config"`
			FunnelWorker              string `json:"funnel-worker"`
			FunnelWorkerServerAddress string `json:"funnel-worker-serveraddress"`
		}
	}
	Project struct {
		ProjectID string `json:"projectId"`
	}
}

func getMetadata(url string) (*metadata, error) {
	meta := &metadata{}
	client := http.Client{}
	path := "/computeMetadata/v1/?recursive=true"
	req, err := http.NewRequest("GET", url+path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Metadata-Flavor", "Google")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Non-200 response status from GCE Metadata: %d", resp.StatusCode)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	log.Debug("RESP", string(body))
	if err != nil {
		return nil, err
	}
	perr := json.Unmarshal(body, meta)
	if perr != nil {
		return nil, perr
	}
	return meta, nil
}
