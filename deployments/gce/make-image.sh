#!/bin/bash

# The name of the VM instance to make an image from.
NAME="funnel-image-$(date +%s)"

# Load helper functions
source "$( dirname $0 )/helpers.sh"

#####################################################

log_header "Checking prerequisites"

if [ ! -e $ROOT/bin/linux_amd64/funnel ]; then
  log_error "Missing ./bin/linux_amd64/funnel binary. Run \`make cross-compile\`"
  exit
fi

if ! check_command gcloud; then
  log_error "Missing 'gcloud' command. https://cloud.google.com/sdk/gcloud/"
  exit
fi

log "Done."

#####################################################

log_header "Creating disk: $NAME"

gce disks create $NAME \
  --image-project ubuntu-os-cloud \
  --image-family ubuntu-1610

#####################################################

log_header "Creating instance: $NAME"

gce instances create $NAME \
  --disk="auto-delete=no,boot=yes,name=$NAME"

#####################################################

# Wait for the instance to start and make ssh available
log_header 'Waiting for instance to start'

gce_wait_for_ssh $NAME

#####################################################

# Make tarball and upload
log_header "Uploading funnel files to $NAME"

gce_ssh $NAME 'mkdir ~/funnel'
gce copy-files $ROOT/bin/linux_amd64/funnel $NAME:~/funnel/
gce copy-files $ROOT/deployments/gce/instance-scripts/start-funnel.sh $NAME:~/funnel/
gce copy-files $ROOT/deployments/gce/instance-scripts/install.sh $NAME:~/funnel/

#####################################################

# Run install script
log_header "Running install.sh"

gce_ssh $NAME 'sudo bash ~/funnel/install.sh'

#####################################################

# Delete image. Disk will remain to be used for the image
log_header "Deleting instance: $NAME"
gce_always --quiet instances delete $NAME 2> /dev/null

#####################################################

# Create an image from the disk
log_header "Creating image: $NAME"
gce images create $NAME --family funnel --source-disk $NAME

#####################################################

log_header "Cleaning up"
gce_always --quiet disks delete $NAME 2> /dev/null

