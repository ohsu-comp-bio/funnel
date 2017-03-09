#!/bin/bash

# deploy-master.sh creates a new GCE instance and starts a funnel master server on it.
# This builds a binary compatible with the GCE OS (linux + amd64).

# Exit on first error
set -e

# Build the code, cross compile to linux + amd64.
GOOS=linux GOARCH=amd64 make

NAME='funnel-master'

#gcloud compute instances create $NAME \
#  --scopes https://www.googleapis.com/auth/cloud-platform \
#  --tags http-server,https-server


RUN="gcloud compute ssh $NAME --command"
COPY="gcloud compute copy-files"

$RUN 'mkdir -p ~/funnel'
$COPY bin/linux_amd64/tes-* share gce $NAME:~/funnel/
$RUN 'sudo mv ~/funnel /opt/'
$RUN 'sudo cp /opt/funnel/gce/funnel-master.service /etc/systemd/system/multi-user.target.wants/'
$RUN 'sudo systemctl daemon-reload'
$RUN 'sudo systemctl start funnel-master.service'
