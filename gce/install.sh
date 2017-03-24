#!/bin/bash

# This runs on the server when it starts (e.g. after the VM is created/restarted).
# This installs the dependencies: docker, gcsfuse, funnel, systemd, etc.
# GCE runs this script as root.
# See: https://cloud.google.com/compute/docs/startupscript

set -o xtrace

# Helper for getting GCE instance metadata
GET_META='curl -f0 -H Metadata-Flavor:Google http://metadata/computeMetadata/v1/instance/attributes'

# Install gcefuse apt repo
GCSFUSE_REPO=gcsfuse-`lsb_release -c -s`
echo "deb http://packages.cloud.google.com/apt $GCSFUSE_REPO main" | sudo tee /etc/apt/sources.list.d/gcsfuse.list
curl https://packages.cloud.google.com/apt/doc/apt-key.gpg | sudo apt-key add -

# Install dependencies
apt update
apt install --yes docker.io gcsfuse

# Create a funnel user/group
useradd -r -s /bin/false funnel
# Add funnel user to the docker group, so it has access to the docker daemon
usermod -aG docker funnel

# The funnel binary and ancillary files are stored in a bucket as a tarball.
# The name/path of that bucket is stored in instance metadata.
# Get that name/path from the metadata
BUNDLE=$( $GET_META/funnel-bundle )

# Download the Funnel bundle and untar it
mkdir -p /opt/funnel
gsutil cp $BUNDLE /opt/funnel/
tar -C /opt/funnel -xzvf /opt/funnel/gce-bundle.tar.gz

# Create the mount point for the shared bucket
# This creates a shared filesystem between the server and workers,
# which is useful for tools like bunny.
mkdir -p /opt/funnel/shared-bucket

# Pull the funnel config from the GCE metadata for this instance.
$GET_META/funnel-config -o /opt/funnel/config.yml

# All files here were created as root, so correct the user/group
chown -R funnel:funnel /opt/funnel
sudo chmod 744 /opt/funnel/gce/start-funnel.sh

# Install the systemd service which keeps the process alive
cp /opt/funnel/gce/funnel.service /etc/systemd/system/multi-user.target.wants/
systemctl daemon-reload
systemctl start funnel.service
