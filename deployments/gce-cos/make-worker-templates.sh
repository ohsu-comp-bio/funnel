#!/bin/bash

# List of machine types to create templates for.
# https://cloud.google.com/compute/docs/machine-types
MACHINE_TYPES="
n1-standard-1
n1-standard-2
n1-standard-4
"

for mt in $MACHINE_TYPES; do
  gcloud compute instance-templates create "funnel-worker-$mt" \
    --scopes compute-rw,storage-rw         \
    --tags funnel                          \
    --image-family cos-stable              \
    --image-project cos-cloud              \
    --machine-type $mt                     \
    --boot-disk-type 'pd-standard'         \
    --boot-disk-size '250GB'               \
    --metadata-from-file user-data=./cloud-init.yaml
done
