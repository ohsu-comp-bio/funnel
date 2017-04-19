#!/bin/bash

# This creates a GCE Instance Template for a Funnel worker.
# The template includes a startup script which installs the worker,
# and can be configured with resources (CPU, RAM, etc).

NAME='funnel-worker-16'
SERVER='funnel-server:9090'
MACHINE_TYPE='n1-standard-16'

gcloud compute instance-templates create $NAME \
  --scopes https://www.googleapis.com/auth/cloud-platform \
  --tags funnel \
  --image-family funnel \
  --machine-type $MACHINE_TYPE \
  --boot-disk-type 'pd-standard' \
  --boot-disk-size '250GB' \
  --metadata "funnel-worker-serveraddress=$SERVER"
