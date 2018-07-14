package papi

import (
	"fmt"

	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/tes"
	"google.golang.org/api/genomics/v2alpha1"
)

// getResources builds the Google Pipelines Resources struct.
func getResources(conf config.GooglePipelines, task *tes.Task) (*genomics.Resources, error) {
	res := task.GetResources()

	if conf.Project == "" {
		return nil, fmt.Errorf("missing project ID")
	}

	// Determine the allowed zones (us-west1-a, us-east1-a, etc)
	// At least one zone is required.
	zones := res.GetZones()
	if len(zones) == 0 {
		if len(conf.DefaultZones) == 0 {
			return nil, fmt.Errorf("at least one zone is required")
		}
		zones = conf.DefaultZones
	}

	machineType, err := pickMachineType(task)
	if err != nil {
		return nil, fmt.Errorf("picking machine type: %v", err)
	}

	return &genomics.Resources{
		ProjectId: conf.Project,
		Zones:     zones,
		VirtualMachine: &genomics.VirtualMachine{
			BootDiskSizeGb: int64(res.GetDiskGb()),
			MachineType:    machineType,
			Preemptible:    res.GetPreemptible(),
		},
	}, nil
}

// pickMachineType tries to pick the Google Cloud machine type that
// best fits the task. Currently it is not optimal.
func pickMachineType(task *tes.Task) (string, error) {
	res := task.GetResources()
	cpus := res.GetCpuCores()
	ram := res.GetRamGb()

	// TODO optimize machine type selection with custom machine types.
	var machineType string
	switch {
	// Standard machine types
	case cpus == 0:
		machineType = "n1-standard-1"
	case cpus == 1 && ram < 3.75:
		machineType = "n1-standard-1"
	case cpus == 2 && ram < 7.5:
		machineType = "n1-standard-2"
	case cpus == 3 || cpus == 4 && ram < 15:
		machineType = "n1-standard-4"
	case cpus >= 5 && cpus <= 8 && ram < 30:
		machineType = "n1-standard-8"
	case cpus >= 9 && cpus <= 16 && ram < 60:
		machineType = "n1-standard-16"
	case cpus >= 17 && cpus <= 32 && ram < 120:
		machineType = "n1-standard-32"
	case cpus >= 33 && cpus <= 64 && ram < 240:
		machineType = "n1-standard-64"
	case cpus > 64 && ram < 360:
		machineType = "n1-standard-96"

	// High-cpu machine types
	case cpus == 2 && ram < 1.8:
		machineType = "n1-highcpu-2"
	case cpus == 3 || cpus == 4 && ram < 3.6:
		machineType = "n1-highcpu-4"
	case cpus >= 5 && cpus <= 8 && ram < 7.2:
		machineType = "n1-highcpu-8"
	case cpus >= 9 && cpus <= 16 && ram < 14.4:
		machineType = "n1-highcpu-16"
	case cpus >= 17 && cpus <= 32 && ram < 28.8:
		machineType = "n1-highcpu-32"
	case cpus >= 33 && cpus <= 64 && ram < 57.6:
		machineType = "n1-highcpu-64"
	case cpus > 64 && ram < 86.4:
		machineType = "n1-highcpu-96"

	// High-mem machine types
	case cpus == 2 && ram < 13:
		machineType = "n1-highmem-2"
	case cpus == 3 || cpus == 4 && ram < 26:
		machineType = "n1-highmem-4"
	case cpus >= 5 && cpus <= 8 && ram < 52:
		machineType = "n1-highmem-8"
	case cpus >= 9 && cpus <= 16 && ram < 104:
		machineType = "n1-highmem-16"
	case cpus >= 17 && cpus <= 32 && ram < 208:
		machineType = "n1-highmem-32"
	case cpus >= 33 && cpus <= 64 && ram < 416:
		machineType = "n1-highmem-64"
	case cpus > 64 && ram < 624:
		machineType = "n1-highmem-96"
	default:
		return "", fmt.Errorf("could not find matching machine type")
	}
	return machineType, nil
}
