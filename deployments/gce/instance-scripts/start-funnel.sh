#!/bin/bash

# This is run on the GCE VM by systemd to start the funnel process.

# Helper for getting GCE instance metadata
GET_META='curl -f0 -H Metadata-Flavor:Google http://metadata/computeMetadata/v1/instance/attributes/'

# Is there any config to download?
CONFIG=$( $GET_META/funnel-config )

if [ "$CONFIG" != "" ]; then
  echo "$CONFIG" > /opt/funnel/funnel.config.yml
fi

# Ensure the config file exists
touch /opt/funnel/funnel.config.yml

# Is this a server or worker?
IS_SERVER=$( $GET_META/funnel-server )
PROCESS="worker"

if [ "$IS_SERVER" == "yes" ]; then
  PROCESS="server"
fi

/opt/funnel/funnel $PROCESS --config /opt/funnel/funnel.config.yml
