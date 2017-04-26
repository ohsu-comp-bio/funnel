#!/bin/bash

# For debug logs
set -o xtrace

# Install dependencies
apt-get update
DEBIAN_FRONTEND=noninteractive apt-get install --yes docker.io iptables-persistent

# Route port 80 to funnel's default port 8000
# TODO if the funnel config changed the port, this won't work
#      could include config during image building and something like
#      funnel config get HTTPPort
iptables -t nat -A PREROUTING -p tcp --dport 80 -j REDIRECT --to-port 8000
# Save the iptables so they persistent on reboot
iptables-save > /etc/iptables/rules.v4

# Create a funnel user/group
useradd -r -s /bin/false funnel
# Add funnel user to the docker group, so it has access to the docker daemon
usermod -aG docker funnel

# Create funnel install dir
mkdir -p /opt/funnel
# Copy the funnel files 
cp ./funnel /opt/funnel/
# All files here were created as root, so correct the user/group
chown -R funnel:funnel /opt/funnel

# Install the systemd service
cp ./funnel.service /etc/systemd/system/multi-user.target.wants/

systemctl daemon-reload
