#!/bin/bash

# This creates a GCE Instance Template for a Funnel worker.
# The template includes a startup script which installs the worker,
# and can be configured with resources (CPU, RAM, etc).

SHARED_BUCKET='smc-rna-funnel'
BUNDLE="gs://$SHARED_BUCKET/gce-bundle.tar.gz"

gcloud compute instance-templates create \
  funnel-worker-16 \
  --scopes https://www.googleapis.com/auth/cloud-platform \
  --tags funnel \
  --machine-type 'n1-standard-16' \
  --boot-disk-type 'pd-standard' \
  --boot-disk-size '250GB' \
  --metadata "funnel-shared-bucket=$SHARED_BUCKET,funnel-bundle=$BUNDLE,funnel-process=worker" \
  --metadata-from-file "startup-script=./gce/install.sh"
