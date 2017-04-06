#!/bin/bash

# The name of the VM instance to make an image from.
NAME="funnel-image-$(date +%s)"

# Directory of this script
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# Load helper functions
source $DIR/helpers.sh

#####################################################

log "Creating disk: $NAME"

gce disks create $NAME \
  --image-project ubuntu-os-cloud \
  --image-family ubuntu-1610

#####################################################

log "Creating instance: $NAME"

gce instances create $NAME \
  --disk="auto-delete=no,boot=yes,name=$NAME"

#####################################################

# Wait for the instance to start and make ssh available
log 'Waiting for instance to start'

gce_wait_for_ssh $NAME

#####################################################

# Make tarball and upload
log "Uploading funnel files to $NAME"

gce_ssh $NAME 'mkdir ~/funnel'
gce copy-files $ROOT/bin/linux_amd64/funnel $NAME:~/funnel/
gce copy-files $ROOT/deployments/gce/instance-scripts/start-funnel.sh $NAME:~/funnel/
gce copy-files $ROOT/deployments/gce/instance-scripts/install.sh $NAME:~/funnel/

#####################################################

# Run install script
log "Running install.sh"

gce_ssh $NAME 'sudo bash ~/funnel/install.sh'

#####################################################

# Delete image. Disk will remain to be used for the image
log "Deleting instance: $NAME"
gce_always --quiet instances delete $NAME 2> /dev/null

#####################################################

# Create an image from the disk
log "Creating image: $NAME"
gce images create $NAME --family funnel --source-disk $NAME

#####################################################

log "Cleaning up"
gce_always --quiet disks delete $NAME 2> /dev/null

