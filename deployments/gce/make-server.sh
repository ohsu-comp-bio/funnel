#!/bin/bash

# This creates a VM on GCE and deploys a Funnel server.
#
# This expects the "funnel" image family to already exist. See make-image.sh
# The default config is "funnel.config.yml" in this directory.

NAME='funnel-server'

# Directory of this script
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# Load helper functions
source $DIR/helpers.sh

log 'Creating VM...'

# Start the VM
gcloud compute instances create $NAME \
  --scopes https://www.googleapis.com/auth/cloud-platform \
  --tags 'funnel,http-server' \
  --image-family funnel \
  --metadata "funnel-server=yes" \
  --metadata-from-file "funnel-instance-config=$DIR/instance-scripts/funnel.config.yml"

# Useful for debugging
gcloud compute instances add-metadata $NAME --metadata=serial-port-enable=1
gcloud compute instances tail-serial-port-output $NAME
