#!/bin/bash

# Name of the VM instance
NAME="funnel-image-builder-$(date +%s)"
# URL of the image installer
INSTALLER_URL='https://github.com/ohsu-comp-bio/funnel/releases/download/dev/funnel-gce-image-installer'

echo "Starting image builder VM instance..."

# Create the VM with a startup script
# which will create the image.
gcloud compute instances create $NAME \
  --scopes compute-rw,storage-rw \
  --zone us-west1-a \
  --tags 'funnel' \
  --metadata "startup-script-url=$INSTALLER_URL,serial-port-enable=1"

# Follow server logs
gcloud compute instances tail-serial-port-output $NAME --zone us-west1-a
