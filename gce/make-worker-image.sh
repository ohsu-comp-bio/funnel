#!/bin/bash

# make-worker-image.sh creates a GCE image from an existing Funnel worker instance.

# The name of the VM instance to make an image from.
SOURCE='funnel-worker'
TS=$(date +%s)
IMAGE="funnel-worker-image-$TS"

function cleanup {
  set +e
  gcloud compute snapshots delete funnel-worker-snapshot-$TS
  gcloud compute disks delete funnel-worker-snapshot-disk-$TS
}
trap cleanup EXIT

# Exit on first error
set -e

echo Creating image: $IMAGE

gcloud compute disks snapshot $SOURCE --snapshot-names funnel-worker-snapshot-$TS

gcloud compute disks create funnel-worker-snapshot-disk-$TS --source-snapshot funnel-worker-snapshot-$TS

gcloud compute images create $IMAGE --source-disk funnel-worker-snapshot-disk-$TS

echo Created image: $IMAGE
