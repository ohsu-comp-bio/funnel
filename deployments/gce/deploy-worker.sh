#!/bin/bash

# Exit on first error
set -e

NAME='funnel-worker'

GOOS=linux GOARCH=amd64 make

gcloud compute instances create $NAME \
  --scopes https://www.googleapis.com/auth/cloud-platform

RUN="gcloud compute ssh $NAME --command"
COPY="gcloud compute copy-files"

$RUN 'mkdir ~/funnel'
$COPY bin/linux_amd64/funnel worker gce $NAME:~/funnel/
$RUN 'sudo mv ~/funnel /opt/'
$RUN 'sudo cp /opt/funnel/gce/funnel-worker.service /etc/systemd/system/multi-user.target.wants/'
$RUN 'sudo apt update'
$RUN 'sudo apt install docker.io'
$RUN 'sudo systemctl daemon-reload'
$RUN 'sudo systemctl start funnel-worker.service'
