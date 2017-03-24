#!/bin/bash

# This creates a VM on GCE and deploys a Funnel server.

NAME='funnel-server'
SHARED_BUCKET='smc-rna-funnel'
BUNDLE="gs://$SHARED_BUCKET/gce-bundle.tar.gz"

# Build the GCE Funnel bundle
#make gce-bundle

# Upload the bundle to the shared bucket
gsutil cp bin/gce-bundle.tar.gz $BUNDLE

# Create rule allowing external HTTP on port 8000 for the Funnel dashboard and HTTP API
gcloud compute firewall-rules create funnel-http-server \
  --allow tcp:8000 \
  --target-tags=funnel-http-server \
  --description='Funnel HTTP server on port 8000'

echo 'Creating VM...'

# Start the VM
gcloud compute instances create $NAME \
  --scopes https://www.googleapis.com/auth/cloud-platform \
  --tags funnel-http-server,funnel \
  --metadata "funnel-shared-bucket=$SHARED_BUCKET,funnel-bundle=$BUNDLE,funnel-process=server" \
  --metadata-from-file "startup-script=./gce/install.sh"

# Useful for debugging
gcloud compute instances add-metadata $NAME --metadata=serial-port-enable=1
gcloud compute instances tail-serial-port-output $NAME
