---
title: Grid Engine
menu:
  main:
    parent: Compute
    weight: 20
---
# Grid Engine

Funnel can be configured to submit workers to [Grid Engine][ge] by making calls
to `qsub`.

The Funnel server needs to run on a submission node.
Configure Funnel to use Grid Engine by including the following config:

```YAML
{{< gridengine-template >}}
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

[ge]: http://gridscheduler.sourceforge.net/documentation.html
