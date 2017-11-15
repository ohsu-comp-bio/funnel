---
title: HTCondor
menu:
  main:
    parent: Compute
    weight: 20
---
# HTCondor

Funnel can be configured to submit workers to [HTCondor][htcondor] by making 
calls to `condor_submit`.

The Funnel server process needs to run on the same machine as the HTCondor master.
Configure Funnel to use HTCondor by including the following config:

```YAML
{{< htcondor-template >}}
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

[htcondor]: https://research.cs.wisc.edu/htcondor/
