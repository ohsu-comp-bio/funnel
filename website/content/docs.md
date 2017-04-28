---
title: Overview
menu:
  main:
    identifier: docs
    weight: -100
---

# Overview

Funnel aims to make batch processing tasks easier to manage by providing a simple
toolkit that supports a wide variety of cluster types. Our goal is to enable you
to spend less time worrying about task management and more time processing data.

## Background

### How Does Funnel Work?

Funnel is a combination of a server and worker processes. First, you define a task.
A task describes input and output files, (Docker) containers and commands, resource
requirements, and some other metadata. You send that task to the Funnel server,
which puts it in the queue until a worker is available. When an appropriate Funnel
worker is available, it downloads the inputs, executes the commands in (Docker)
containers, and uploads the outputs.

Funnel also comes with some tools related to managing workers and tasks. There's
a dashboard, a scheduler, an autoscaler, some rudimentary workflow tools, and more.

### Why Does Funnel Exist?

Here at OHSU Computational Biology, a typical project involves coordinating dozens
of tasks across hundreds of CPUs in a cluster of machines in order to process hundreds
of files. That's standard fare for most computational groups these days, and for some
groups it's "thousands" or "millions" instead of "hundreds".

Because we're part of a worldwide scientific community, it's important that we're able
to easily share our work. If we create a variant calling pipeline with 50 steps,
we need people outside OHSU to run that pipeline easily and efficiently.

There's a long list of projects making great strides in the tools we use to tackle
this type of work, but they have a common problem. Every group of users has grown
a different set of tools for managing and interacting with their cluster. Some use
HTCondor and NFS. Some use Open Grid Engine and Lustre. Some prefer cloud providers,
but which one? Google? Amazon? Each cluster comes with a different interface to learn
(and a new set of problems to debug too).

Tool authors usually end up writing (and hopefully maintaining) a set of
compute and storage plugins for each type of cluster. Many authors don't have
time for that, and their tools end up being limited to their environment.
Some tools were never meant to be shared, instead they were originally just
a prototype or a set of helper scripts for working with AWS instances.

The [GA4GH Task Execution Schemas][tes] (TES) group aims to ease problems by
designing a simple API for data processing tasks that can be easily layered on top of,
or easily plugged into, most existing cluster. Funnel started as the first
implementation of the TES API.

Funnel aims to ease these problems. Our goal is to enable easy management of tasks
and tools that need to work across many types of clusters.

[galaxy]: https://galaxyproject.org/
[cwl]: http://commonwl.org/
[wdl]: https://software.broadinstitute.org/wdl/
[tes]: https://github.com/ga4gh/task-execution-schemas
