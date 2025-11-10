---
title: GCP Batch
menu:
  main:
    parent: Compute
    weight: 20
---

# GCP Batch

This guide covers deploying a Funnel server that leverages [Google Cloud Batch][0] for task execution. GCP Batch is a fully managed service that lets you schedule, queue, and execute batch computing workloads on Google Cloud Platform.

## Setup

Get started by ensuring you have the necessary Google Cloud Platform resources configured:

1. A GCP project with the Batch API enabled
2. Appropriate IAM permissions to create and manage Batch jobs
3. Service account credentials (optional - can use Application Default Credentials)

The Funnel GCP Batch backend will automatically use [Google Application Default Credentials][1]
if no explicit credentials are provided. This means it can authenticate using:
- Service account keys
- gcloud CLI authentication
- Workload Identity (when running on GKE)
- Compute Engine service accounts (when running on GCE)

### Steps

* [Enable the Batch API][2] in your GCP project
* [Configure IAM permissions][3] for the service account that will run Funnel
* Choose a [GCP region][4] for job execution (e.g., `us-central1`)
* (Optional) Configure a database backend for storing task metadata (e.g., Datastore, MongoDB)

For more information, check out the [GCP Batch documentation][0].

### Prerequisites

Ensure your service account or user has the following IAM roles:
- `roles/batch.jobsEditor` - To create and manage Batch jobs
- `roles/iam.serviceAccountUser` - To run jobs as a service account
- Storage permissions (e.g., `roles/storage.objectAdmin`) - For input/output file handling

## Configuring the Funnel Server

Below is an example configuration. Note that credentials can be left blank,
allowing Funnel to automatically load credentials from the environment using
Application Default Credentials.

```yaml
Compute: "gcp-batch"

GCPBatch:
  Project: "example-project-id"
  Location: "us-central1"

  # Reconciler for task reporting
  DisableReconciler: false
  ReconcileRate: 10m

# Google Cloud Storage configuration
GoogleStorage:
  Disabled: false
  # Optional: Path to service account credentials
  # If not provided, will use Application Default Credentials
  CredentialsFile: ""
```

### Configuration Options

- **Project** (required): Your GCP project ID
- **Location** (required): The GCP region where jobs will be executed (e.g., `us-central1`, `europe-west1`)
- **DisableReconciler** (optional): When `false`, Funnel periodically checks job states with GCP Batch to detect stuck or failed jobs. Default: `false`
- **ReconcileRate** (optional): How often to reconcile task states. Default: `10m`

### Start the server

```sh
funnel server run --config /path/to/config.yaml
```

Or use environment variables for configuration:

```sh
export FUNNEL_COMPUTE=gcp-batch
export FUNNEL_GCPBATCH_PROJECT=my-project-id
export FUNNEL_GCPBATCH_LOCATION=us-central1

funnel server run
```

### Known issues

- Task cancellation is not yet fully implemented
- State reconciliation is not yet fully implemented
- Backend parameters support is in development

## How it works

When a task is submitted to Funnel:

1. Funnel translates the TES (Task Execution Schema) task into a GCP Batch job
2. The job is submitted to the GCP Batch API
3. GCP Batch schedules and executes the job on managed compute resources
4. Funnel's reconciler (if enabled) periodically checks job status
5. Task outputs and logs are stored according to the configured storage backend

[0]: https://cloud.google.com/batch/docs
[1]: https://cloud.google.com/docs/authentication/application-default-credentials
[2]: https://console.cloud.google.com/apis/library/batch.googleapis.com
[3]: https://cloud.google.com/batch/docs/get-started#grant-access
[4]: https://cloud.google.com/batch/docs/locations
