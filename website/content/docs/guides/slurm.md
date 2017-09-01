---
title: SLURM

menu:
  main:
    parent: guides
    weight: 20
---

# SLURM

Funnel can be configured to submit workers to [SLURM][slurm] by making calls
to `sbatch`.

The Funnel server process needs to run on the same machine as the SLURM master.  
Configure Funnel to use SLURM by including the following config:

```YAML
{{< slurm-template >}}
```
The following variables are available for use in the template:

| Variable    |  Description |
|:------------|:-------------|
|TaskId       | funnel task id |
|Executable   | path to funnel binary |
|Config       | path to the funnel worker config file |
|WorkDir      | funnel working directory |
|Cpus         | requested cpu cores |
|RamGb        | requested ram |
|DiskGb       | requested free disk space |
|Zone         | requested zone (could be used for queue name) |
|Project      | project (could be used for account to charge) |

See https://golang.org/pkg/text/template for information on creating templates.

[slurm]: https://slurm.schedmd.com/
