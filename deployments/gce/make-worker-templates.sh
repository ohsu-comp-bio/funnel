#!/bin/bash

# This creates a GCE Instance Templates for a Funnel workers.
# The template includes a startup script which installs the worker,
# and can be configured with resources (CPU, RAM, etc).

SERVER='funnel-server:9090'
MACHINE_TYPES="
n1-standard-1
n1-standard-8
n1-standard-16
"

for mt in $MACHINE_TYPES; do
  NAME="funnel-worker-$mt"
  gcloud compute instance-templates create $NAME \
    --scopes https://www.googleapis.com/auth/cloud-platform \
    --tags funnel \
    --image-family funnel \
    --machine-type $mt \
    --boot-disk-type 'pd-standard' \
    --boot-disk-size '250GB' \
    --metadata "funnel-worker-serveraddress=$SERVER"
done
