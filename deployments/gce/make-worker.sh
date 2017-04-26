#!/bin/bash

# This creates a VM on GCE and deploys a Funnel worker.
#
# This expects a Funnel server to already be running,
# and also expects the "funnel" image family to already exist.

NAME="funnel-worker-$(date +%s)"
FUNNEL_SERVER='funnel-server:9090'

gcloud compute instances create $NAME \
  --scopes compute-rw,storage-rw \
  --zone 'us-west1-a' \
  --tags funnel \
  --image-family funnel \
  --machine-type 'n1-standard-16' \
  --boot-disk-type 'pd-standard' \
  --boot-disk-size '250GB' \
  --metadata "funnel-worker-serveraddress=$FUNNEL_SERVER"

# Tail serial port logs.
# Useful for debugging.
gcloud compute instances add-metadata $NAME --metadata=serial-port-enable=1
gcloud compute instances tail-serial-port-output $NAME
