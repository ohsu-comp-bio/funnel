package gce

import (
	"context"
	"fmt"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/compute/v1"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"tes/config"
	pbr "tes/server/proto"
)

type gceClientI interface {
	// Templates returns available worker types,
	// which are described using GCE instance templates.
	Templates() []*pbr.Worker
	StartWorker(*pbr.Worker) error
}

type gceClient struct {
	templates []*pbr.Worker
	conf      config.Config
	svc       *compute.Service
}

// service creates a Google Compute service client
func newGCEClient(ctx context.Context, conf config.Config) (gceClientI, error) {
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

	return &gceClient{
		conf: conf,
		svc:  svc,
	}, nil
}

// Templates queries the GCE API to get details about GCE instance templates.
// If the API client fails to connect, this returns an empty list.
func (s *gceClient) Templates() []*pbr.Worker {
	// templates are cached after the first call
	if s.templates != nil {
		return s.templates
	}

	templates := []*pbr.Worker{}
	project := s.conf.Schedulers.GCE.Project
	machineTypes := map[string]pbr.Resources{}

	// Get the machine types available to the project + zone
	resp, err := s.svc.MachineTypes.List(project, s.conf.Schedulers.GCE.Zone).Do()
	if err != nil {
		log.Error("Couldn't get GCE machine list.")
		return templates
	}

	for _, m := range resp.Items {
		machineTypes[m.Name] = pbr.Resources{
			Cpus: uint32(m.GuestCpus),
			Ram:  float64(m.MemoryMb) / float64(1024),
		}
	}

	for _, t := range s.conf.Schedulers.GCE.Templates {
		// Get the instance template from the GCE API
		tpl, err := s.svc.InstanceTemplates.Get(project, t).Do()
		if err != nil {
			log.Error("Couldn't get GCE template. Skipping.", "error", err, "template", t)
			continue
		}
		// Map the machine type ID string to a pbr.Resources struct
		res := machineTypes[tpl.Properties.MachineType]
		// TODO is there always at least one disk? Is the first the best choice?
		//      how to know which to pick if there are multiple?
		res.Disk = float64(tpl.Properties.Disks[0].InitializeParams.DiskSizeGb)
		templates = append(templates, &pbr.Worker{
			Resources: &res,
		})
	}

	// Looks like we have a successful response (no errors above)
	// so cache the templates array.
	s.templates = templates
	return templates
}

// StartWorker calls out to GCE APIs to create a VM instance
// with a Funnel worker.
func (s *gceClient) StartWorker(w *pbr.Worker) error {
	project := s.conf.Schedulers.GCE.Project
	zone := s.conf.Schedulers.GCE.Zone

	// Get the instance template from the GCE API
	tpl, terr := s.svc.InstanceTemplates.Get(project, w.Gce.Template).Do()
	if terr != nil {
		log.Error("Couldn't retrieve GCE instance template",
			"error", terr,
			"template", w.Gce.Template)
		return terr
	}

	c := s.conf.Worker
	c.ID = w.Id
	c.Timeout = -1
	c.Storage = s.conf.Storage
	workdir := path.Join(s.conf.WorkDir, w.Id)
	workdir, _ = filepath.Abs(workdir)
	os.MkdirAll(workdir, 0755)
	confYaml := string(c.ToYaml())

	// Add the funnel config yaml string to the template metadata
	props := tpl.Properties
	metadata := compute.Metadata{
		Items: append(props.Metadata.Items,
			&compute.MetadataItems{
				Key:   "funnel-config",
				Value: &confYaml,
			},
		),
	}

	// Prepare disk details by setting the specific zone
	for _, disk := range props.Disks {
		dt := localize(zone, "diskTypes", disk.InitializeParams.DiskType)
		disk.InitializeParams.DiskType = dt
	}

	instance := compute.Instance{
		Name:              w.Id,
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

	op, ierr := s.svc.Instances.Insert(project, zone, &instance).Do()
	if ierr != nil {
		log.Error("Couldn't insert GCE VM instance", ierr)
		return ierr
	}

	log.Debug("GCE VM instance created", "details", op)
	return nil
}

func localize(zone, resourceType, val string) string {
	return fmt.Sprintf("zones/%s/%s/%s", zone, resourceType, val)
}
