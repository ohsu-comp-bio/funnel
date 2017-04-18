#!/bin/bash

# This creates a VM on GCE and deploys a Funnel worker.
# This expects a Funnel server to already be running,
# and also expects the "funnel" image family to already exist.

NAME="funnel-worker-$(date +%s)"
SERVER='funnel-server:9090'
MACHINE_TYPE='n1-standard-16'

gcloud compute instances create $NAME \
  --scopes https://www.googleapis.com/auth/cloud-platform \
  --tags funnel \
  --image-family funnel \
  --machine-type $MACHINE_TYPE \
  --boot-disk-type 'pd-standard' \
  --boot-disk-size '250GB' \
  --metadata "funnel-worker-serveraddress=$SERVER"

# Tail serial port logs.
# Useful for debugging.
gcloud compute instances add-metadata $NAME --metadata=serial-port-enable=1
gcloud compute instances tail-serial-port-output $NAME
