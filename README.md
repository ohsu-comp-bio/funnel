![master-build-status](https://travis-ci.org/ohsu-comp-bio/funnel.svg?branch=master)
[![Gitter](https://badges.gitter.im/ohsu-comp-bio/funnel.svg)](https://gitter.im/ohsu-comp-bio/funnel?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge)

Funnel (alpha)
======

Funnel is a server for executing tasks on a cluster. Given a task description, Funnel will find a worker to execute the task, download inputs, run a series of (Docker) containers, upload outputs, capture logs, and track the whole process.

Funnel is an implementation of the [GA4GH Task Execution Schemas](https://github.com/ga4gh/task-execution-schemas), an effort to standardize the APIs used for task execution across many platforms.

Funnel provides an API server, multiple storage backends (local FS, S3, Google Bucket, etc.), multiple compute backends (local, HTCondor, Google Cloud, etc.), and a web dashboard.

See the other docs pages for more detail:

- [Getting Started](./docs/getting-started.md)
- [Development](./docs/development.md)
- [APIs](./docs/apis.md)
- [Godocs](https://godoc.org/github.com/ohsu-comp-bio/funnel)
- [Design](./docs/design.md)
- Guides
  - [Google Cloud Compute](./docs/guides/google-cloud-compute.md)
  - [HTCondor](./docs/guides/htcondor.md)
