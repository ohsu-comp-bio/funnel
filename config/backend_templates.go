package config

// The following variables are available for use in the templates:
//
// WorkerId       funnel worker id
// Executable     path to funnel binary
// WorkerConfig   path to the funnel worker config file
// WorkDir        funnel working directory
// Cpus           requested cpus
// RamGb          requested ram
// DiskGb         requested free disk space

// See https://golang.org/pkg/text/template for more information

var slurmTemplate = `#!/bin/bash
#SBATCH --job-name {{.WorkerId}}
#SBATCH --ntasks 1
#SBATCH --error {{.WorkDir}}/funnel-worker-stderr
#SBATCH --output {{.WorkDir}}/funnel-worker-stdout
{{if ne .Cpus 0 -}}
{{printf "#SBATCH --cpus-per-task %d" .Cpus}}
{{- end}}
{{if ne .RamGb 0.0 -}}
{{printf "#SBATCH --mem %.0fGB" .RamGb}}
{{- end}}
{{if ne .DiskGb 0.0 -}}
{{printf "#SBATCH --tmp %.0fGB" .DiskGb}}
{{- end}}

{{.Executable}} worker --config {{.WorkerConfig}}
`

var condorTemplate = `
universe       = vanilla
environment    = "PATH=/usr/bin"
executable     = {{.Executable}}
arguments      = worker --config worker.conf.yml
log            = {{.WorkDir}}/condor-event-log
error          = {{.WorkDir}}/funnel-worker-stderr
output         = {{.WorkDir}}/funnel-worker-stdout
input          = {{.WorkerConfig}}

should_transfer_files   = YES
when_to_transfer_output = ON_EXIT

{{if ne .Cpus 0 -}}
{{printf "request_cpus = %d" .Cpus}}
{{- end}}
{{if ne .RamGb 0.0 -}}
{{printf "request_memory = %.2f GB" .RamGb}}
{{- end}}
{{if ne .DiskGb 0.0 -}}
{{printf "request_disk = %.2 fGB" .DiskGb}}
{{- end}}

queue
`
