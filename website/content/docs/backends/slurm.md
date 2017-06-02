---
title: SLURM

menu:
  main:
    parent: backends
    weight: 20
---

# SLURM

Funnel can be configured to start workers via [SLURM][slurm].  

Workers will start, execute a single task, then exit immediately so that they don't
unfairly hold slots in the SLURM queue. Funnel accesses SLURM by making calls
to `sbatch`.

The Funnel server process needs to run on the same machine as the SLURM master.  
Configure Funnel to use SLURM by including the following config:

```YAML
{{< slurm-template >}}
```
The following variables are available for use in the template:

| Variable    |  Description |
|:------------|:-------------|
|WorkerId     |  funnel worker id |
|Executable   |  path to funnel binary |
|WorkerConfig |  path to the funnel worker config file |
|WorkDir      |  funnel working directory |
|Cpus         |  requested cpu cores |
|RamGb        | requested ram |
|DiskGb       | requested free disk space |
|Zone         | requested zone (could be used for queue name) |
|Project      | project (could be used for account to charge) |

See https://golang.org/pkg/text/template for information on creating templates.

To start Funnel with a config file:
```shell
$ funnel server --config ./your-config.yaml
```

[slurm]: https://slurm.schedmd.com/
