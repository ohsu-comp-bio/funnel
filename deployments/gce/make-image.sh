#!/bin/bash

# Name of the VM
NAME='buchanan-test'
# URL of the image installer
INSTALLER_URL='gs://smc-rna-funnel/bundle.run'

# Create the VM with a startup script
# which will create the image.
gcloud compute instances create $NAME \
  --scopes compute-rw \
  --metadata "startup-script-url=$INSTALLER_URL,serial-port-enable=1"

# Follow server logs
gcloud compute instances tail-serial-port-output $NAME
