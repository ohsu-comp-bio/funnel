---
title: Slurm
menu:
  main:
    parent: Compute
    weight: 20
---
# Slurm

Funnel can be configured to submit workers to [Slurm][slurm] by making calls
to `sbatch`.

The Funnel server needs to run on a submission node.
Configure Funnel to use Slurm by including the following config:

```YAML
{{< slurm-template >}}
```
The following variables are available for use in the template:

| Variable    |  Description |
|:------------|:-------------|
|TaskId       | funnel task id |
|WorkDir      | funnel working directory |
|Cpus         | requested cpu cores |
|RamGb        | requested ram |
|DiskGb       | requested free disk space |
|Zone         | requested zone (could be used for queue name) |

See https://golang.org/pkg/text/template for information on creating templates.

[slurm]: https://slurm.schedmd.com/
