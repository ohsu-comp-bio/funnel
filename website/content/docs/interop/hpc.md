---
title: HPC
menu:
  main:
    parent: Interop
---

> ⚠️ HPC support is in active development and commands may change.

# HPC

## Getting Started

Funnel supports submitting tasks to HPC (High-Performance Computing) backends via job schedulers like Slurm.

To get started with using HPC with Funnel, you will need to configure Funnel to use a suitable container engine and job scheduler. This guide will walk you through setting up the necessary configurations and running tasks on an HPC system using Slurm.

## Example

Below is an example configuration for setting up Funnel to run tasks on an HPC system using Slurm as the job scheduler and exadocker as the container engine. This configuration includes settings for local storage, worker behavior, server details, and Slurm job submission templates.

<details>
  <summary>config.yml</summary>

```yaml
LocalStorage:
  # Whitelist of local directory paths which Funnel is allowed to access.
  AllowedDirs:
    - ./
    - /home/users/example

Worker:
  LeaveWorkDir: true
  ContainerDriver: sudo /opt/acc/sbin/exadocker

Server:
  HostName: exahead1

Compute: slurm

# Custom template for Slurm
Slurm:
  Template: |
    #!/bin/bash
    #SBATCH --job-name {{.TaskId}}
    #SBATCH --ntasks 1
    #SBATCH --error {{.WorkDir}}/funnel-stderr
    #SBATCH --output {{.WorkDir}}/funnel-stdout
    #SBATCH --gres disk:1
    {{if ne .Cpus 0 -}}
    {{printf "#SBATCH --cpus-per-task %d" .Cpus}}
    {{- end}}
    {{if ne .RamGb 0.0 -}}
    {{printf "#SBATCH --mem %.0fGB" .RamGb}}
    {{- end}}
    {{if ne .DiskGb 0.0 -}}
    {{printf "#SBATCH --tmp %.0fGB" .DiskGb}}
    {{- end}}

    srun /usr/local/bin/mkdir-scratch.sh > /dev/null
    SCRATCH_PATH="/mnt/scratch/${SLURM_JOB_ID}"

    cd $SCRATCH_PATH
    srun bash -c 'funnel worker run --taskID {{.TaskId}} {{.Args}} --Worker.ScratchPath .' --temp-dir $SCRATCH_PATH

    cd - > /dev/null
    srun /usr/local/bin/rmdir-scratch.sh > /dev/null
```
</details>

<details>
  <summary>md5sum.json</summary>

```json
{
  "name": "md5sum example",
  "description": "Demonstrates input and output files using a simple md5sum command.",
  "inputs": [
    {
      "name": "md5sum input",
      "description": "Input to md5sum. /tmp/md5sum_input must exist on the host system.",
      "url": "file:///tmp/md5sum_input",
      "type": "FILE",
      "path": "/tmp/in"
    }
  ],
  "outputs": [
    {
      "name": "md5sum stdout",
      "description": "Stdout of md5sum is captured and copied to /tmp/md5_output on the host system.",
      "url": "file:///tmp/md5sum_output",
      "type": "FILE",
      "path": "/tmp/out"
    }
  ],
  "executors": [
    {
      "image": "alpine",
      "command": ["md5sum", "/tmp/in"],
      "stdout": "/tmp/out"
    }
  ]
}
```
</details>

## Start Funnel Server

```sh
funnel server run --config config.yml
```

## Submit Task

```sh
funnel task create md5sum.json
```