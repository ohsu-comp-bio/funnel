#!/bin/bash

# This creates a VM on GCE and deploys a Funnel server.
#
# This expects the "funnel" image family to already exist.

NAME='funnel-server'

#gcloud compute firewall-rules create default-http --allow='tcp:80'
#gcloud compute firewall-rules create default-http --allow='tcp:80' --source-tags='http-server'

# Start the VM
gcloud compute instances create $NAME    \
  --scopes       'compute-rw,storage-rw' \
  --zone         'us-west1-a'            \
  --tags         'funnel,http-server'    \
  --image-family 'funnel'

# Useful for debugging
gcloud compute instances add-metadata $NAME --metadata=serial-port-enable=1 --zone us-west1-a
gcloud compute instances tail-serial-port-output $NAME --zone us-west1-a
