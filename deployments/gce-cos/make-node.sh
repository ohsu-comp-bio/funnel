#!/bin/bash

# This creates a VM on GCE and deploys a Funnel node.
#
# This expects a Funnel server to already be running.

NAME="funnel-node-$(date +%s)"
FUNNEL_SERVER='funnel-server:9090'

gcloud compute instances create $NAME                        \
  --scopes compute-rw,storage-rw                             \
  --zone 'us-west1-a'                                        \
  --tags funnel                                              \
  --image-family cos-stable                                  \
  --image-project cos-cloud                                  \
  --machine-type 'n1-standard-4'                             \
  --boot-disk-type 'pd-standard'                             \
  --boot-disk-size '250GB'                                   \
  --metadata "funnel-node-serveraddress=$FUNNEL_SERVER"    \
  --metadata-from-file user-data=./cloud-init.yaml

# Tail serial port logs.
# Useful for debugging.
#gcloud compute instances add-metadata $NAME --metadata=serial-port-enable=1
#gcloud compute instances tail-serial-port-output $NAME
