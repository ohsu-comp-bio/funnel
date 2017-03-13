#!/bin/bash

# make-worker-image.sh creates a GCE image from an existing Funnel worker instance.

# The name of the VM instance to make an image from.
SOURCE='funnel-worker'

# Exit on first error
set -e

gcloud compute disks snapshot $SOURCE --snapshot-names funnel-worker-snapshot

gcloud compute disks create funnel-worker-image-snapshot --source-snapshot funnel-worker-snapshot

gcloud compute images create funnel-worker-image --source-disk funnel-worker-image-snapshot

gcloud compute disks delete funnel-worker-image-snapshot

gcloud compute snapshots delete funnel-worker-snapshot
