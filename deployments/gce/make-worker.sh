#!/bin/bash

# This creates a VM on GCE and deploys a Funnel worker.

NAME='funnel-worker'

# Directory of this script
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# Load helper functions
source $DIR/helpers.sh

log_header 'Creating VM...'

gcloud compute instances create $NAME \
  --scopes https://www.googleapis.com/auth/cloud-platform \
  --tags funnel \
  --image-family funnel \
  --metadata-from-file "funnel-instance-config=$DIR/instance-scripts/funnel.config.yml"

# Useful for debugging
gcloud compute instances add-metadata $NAME --metadata=serial-port-enable=1
gcloud compute instances tail-serial-port-output $NAME
