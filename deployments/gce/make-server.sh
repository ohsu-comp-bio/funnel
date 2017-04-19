#!/bin/bash

# This creates a VM on GCE and deploys a Funnel server.
#
# This expects the "funnel" image family to already exist. See make-image.sh

NAME='funnel-server'

# Start the VM
gcloud compute instances create $NAME \
  --scopes https://www.googleapis.com/auth/cloud-platform \
  --tags 'funnel,http-server' \
  --image-family funnel

# Useful for debugging
gcloud compute instances add-metadata $NAME --metadata=serial-port-enable=1
gcloud compute instances tail-serial-port-output $NAME
