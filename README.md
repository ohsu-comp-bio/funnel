[![Build Status](https://img.shields.io/github/actions/workflow/status/ohsu-comp-bio/funnel/tests.yaml)](https://github.com/ohsu-comp-bio/funnel/actions/workflows/tests.yaml)
[![Compliance Tests Status](https://img.shields.io/github/actions/workflow/status/ohsu-comp-bio/funnel/compliance-test.yaml?label=Compliance%20Tests)](https://github.com/ohsu-comp-bio/funnel/actions/workflows/compliance-test.yaml)
[![Gitter](https://badges.gitter.im/ohsu-comp-bio/funnel.svg)](https://gitter.im/ohsu-comp-bio/funnel)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Godoc](https://img.shields.io/badge/godoc-ref-blue.svg)](http://godoc.org/github.com/ohsu-comp-bio/funnel)


Funnel
======

https://ohsu-comp-bio.github.io/funnel/

Funnel is a toolkit for distributed, batch task execution, including a server, worker, and a set of compute, storage, and database backends. Given a task description, Funnel will find a worker to execute the task, download inputs, run a series of (Docker) containers, upload outputs, capture logs, and track the whole process.

Funnel is an implementation of the [GA4GH Task Execution Schemas](https://github.com/ga4gh/task-execution-schemas), an effort to standardize the APIs used for task execution across many platforms.

Funnel provides an API server, multiple storage backends (local FS, S3, Google Bucket, etc.), multiple compute backends (local, HTCondor, Google Cloud, etc.), and a web dashboard.
