#!/bin/bash

# This creates a VM on GCE and deploys a Funnel worker.

NAME='funnel-worker'
SHARED_BUCKET='smc-rna-funnel'
BUNDLE="gs://$SHARED_BUCKET/gce-bundle.tar.gz"

# Build the GCE Funnel bundle
#make gce-bundle

# Upload the bundle to the shared bucket
#gsutil cp bin/gce-bundle.tar.gz $BUNDLE

gcloud compute instances create $NAME \
  --scopes https://www.googleapis.com/auth/cloud-platform \
  --tags funnel \
  --metadata "funnel-shared-bucket=$SHARED_BUCKET,funnel-bundle=$BUNDLE,funnel-process=worker" \
  --metadata-from-file startup-script=gce/install.sh

# Useful for debugging
gcloud compute instances add-metadata $NAME --metadata=serial-port-enable=1
gcloud compute instances tail-serial-port-output $NAME
