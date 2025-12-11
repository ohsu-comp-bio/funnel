---
title: GCP Batch
menu:
  main:
    parent: Compute
    weight: 20
---

> [!WARNING]
>
> GCP Batch support is in early development.
>
> Production deployments are recommended to use the alternative compute backends available, such as AWS Batch, Kubernetes, Slurm, etc.

# Current State

- [x] Submitting tasks with Input + Output via single Google Storage bucket (e.g. [example job](https://console.cloud.google.com/batch/jobsDetail/regions/us-central1/jobs/d49rl1q9io6s73erbng0/details?authuser=1&project=tes-batch-integration-test) + [tes-batch-integration](https://console.cloud.google.com/storage/browser/tes-batch-integration))

<details>
  <summary>Example TES Task</summary>

```json
{
  "name": "Input/Output Test",
  "inputs": [{
    "url": "gs://tes-batch-integration/README.md",
    "path": "/mnt/disks/tes-batch-integration/README.md"
  }],
  "outputs": [{
    "url": "gs://tes-batch-integration/README.md.sha256",
    "path": "/mnt/disks/tes-batch-integration/README.md.sha256"
  }],
  "executors": [{
    "image": "alpine",
    "command": ["sha256sum", "/mnt/disks/tes-batch-integration/README.md | tee /mnt/disks/tes-batch-integration/README.md.sha256"]
  }]
}
```

</details>

# Limitations

- [ ] Task State Syncing: need to update reconciler to fetch Task State ([Issue](https://github.com/ohsu-comp-bio/funnel/issues/1270))
- [ ] Task Logs: appear in GCP Batch Console, but not retrieved by Funnel yet ([Issue](https://github.com/ohsu-comp-bio/funnel/issues/1271))
- [ ] Test against multiple buckets ([Issue](https://github.com/ohsu-comp-bio/funnel/issues/1272))
- [ ] Verify Compliance Tests ([Issue](https://github.com/ohsu-comp-bio/funnel/issues/1273))

---

# Overview

Following are the steps to install, configure, and start the Funnel server and submit an example task.

# Quick Start

## 1. Install Funnel

```sh
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/ohsu-comp-bio/funnel/refs/heads/develop/install.sh)" -- v0.11.7-rc.10
```

## 2. Start Server

<details>
  <summary><code>Config Example</code></summary>

```yaml
Compute: gcp-batch

GCPBatch:
  Project: example-project
  Location: us-central1
```

</details>

```sh
funnel server run --Compute "gcp-batch" --GCPBatch.Project "example-project" --GCPBatch.Location "us-central1"
```

## 3. Submit Task

<details>
  <summary><code>gcp-example.json</code></summary>

```json
{
  "name": "Input/Output Test",
  "inputs": [
    {
      "url": "gs://tes-batch-integration/README.md",
      "path": "/mnt/disks/tes-batch-integration/README.md"
    }
  ],
  "outputs": [
    {
      "url": "gs://tes-batch-integration/README.md.sha256",
      "path": "/mnt/disks/tes-batch-integration/README.md.sha256"
    }
  ],
  "executors": [
    {
      "image": "alpine",
      "command": [
        "sha256sum",
        "/mnt/disks/tes-batch-integration/README.md | tee /mnt/disks/tes-batch-integration/README.md.sha256"
      ]
    }
  ]
}
```

</details>

```sh
funnel task create gcp-example.json
<TASK ID>
```

## 4. Query Task

```
funnel task get <TASK ID>
```

# Additional Resources

- https://docs.cloud.google.com/batch/docs
