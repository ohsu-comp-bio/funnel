---
title: GCP Batch
menu:
  main:
    parent: Compute
    weight: 20
---

# Overview

Following are the steps to install, configure, and start the Funnel server and submit an example task.

# Quick Start

## Install Funnel

```sh
➜ /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/ohsu-comp-bio/funnel/refs/heads/develop/install.sh)" -- v0.11.7-rc.10
```

## Configure Server

`config.yaml`
```yaml
Compute: gcp-batch

GCPBatch:
  DisableReconciler: True
  ReconcileRate: 10s
  Project: tes-batch-integration-test
  Location: us-central1

GoogleStorage:
  Disabled: false
```

## Start Server

```sh
➜ funnel server run --config config.yaml
```

## Submit Task

`gcp-example.json`
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

```sh
➜ funnel task create gcp-example.json
```

# Additional Resources

- https://docs.cloud.google.com/batch/docs
