#!/bin/bash

# This creates a VM on GCE and deploys a Funnel server.

NAME='funnel-server'

# Ensure that a firewall rule exists allowing HTTP traffic
gcloud compute firewall-rules create default-http --allow='tcp:80' --source-tags='http-server' --quiet

# Start the VM
gcloud compute instances create $NAME    \
  --scopes       'compute-rw,storage-rw' \
  --zone         'us-west1-a'            \
  --tags         'funnel,http-server'    \
  --machine-type n1-standard-2           \
  --image-family cos-stable              \
  --image-project cos-cloud              \
  --metadata-from-file user-data=./cloud-init.yaml

# Useful for debugging
#gcloud compute instances add-metadata $NAME --metadata=serial-port-enable=1 --zone us-west1-a
#gcloud compute instances tail-serial-port-output $NAME --zone us-west1-a
