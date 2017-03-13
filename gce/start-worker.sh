#!/bin/bash

# start-worker.sh executes on the worker machine in GCE (probably via systemd)
# this starts the worker.

# Pull the funnel worker config yaml from the GCE metadata for this instance.
# This allows the funnel scheduler the flexibility to configure workers differently.
curl -o /opt/funnel/gce-metadata-worker.config.yml -f0 -H Metadata-Flavor:Google http://metadata/computeMetadata/v1/instance/attributes/funnel-config

if [ -f /opt/funnel/gce-metadata-worker.config.yml ]; then
  CONF=/opt/funnel/gce-metadata-worker.config.yml
else
  # GCE metadata config didn't exist, so fallback to default config
  CONF=/opt/funnel/gce/default-worker.config.yml
fi

/opt/funnel/tes-worker -config $CONF
