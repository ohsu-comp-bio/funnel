#!/bin/bash

curl -H Metadata-Flavor:Google http://metadata/computeMetadata/v1/instance/attributes/funnel-config > /opt/funnel/funnel.config.yml

/opt/funnel/bin/tes-worker -config /opt/funnel/funnel.config.yml
