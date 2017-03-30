#!/bin/bash

# This is run on the GCE VM by systemd to start the funnel process.
# Related: gce/funnel.service

# Helper for getting GCE instance metadata
GET_META='curl -f0 -H Metadata-Flavor:Google http://metadata/computeMetadata/v1/instance/attributes/'

# Get the name of the shared bucket from GCE metadata
SHARED_BUCKET=$( $GET_META/funnel-shared-bucket )

# Mount the shared bucket
# This creates a shared filesystem between the server and workers,
# which is useful for tools like bunny.
gcsfuse $SHARED_BUCKET /opt/funnel/shared-bucket

# Is this a server or worker?
PROCESS=$( $GET_META/funnel-process )

# Start the funnel process
if [ "$PROCESS" == "server" ]; then
  # If there was no config in the metadata, copy over the default config
  cp -n /opt/funnel/gce/default-server.config.yml /opt/funnel/config.yml
  /opt/funnel/bin/linux_amd64/funnel server --config /opt/funnel/config.yml
else
  # If there was no config in the metadata, copy over the default config
  cp -n /opt/funnel/gce/default-worker.config.yml /opt/funnel/config.yml
  /opt/funnel/bin/linux_amd64/funnel worker --config /opt/funnel/config.yml
fi
