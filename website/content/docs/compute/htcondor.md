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

The Funnel server needs to run on a submission node.
Configure Funnel to use HTCondor by including the following config:

It is recommended to update the submit file template so that the
`funnel worker run` command takes a config file as an argument 
(e.g. `funnel worker run --config /opt/funnel_config.yml --taskID {{.TaskId}}`)

```YAML
{{< htcondor-template >}}
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

[htcondor]: https://research.cs.wisc.edu/htcondor/
