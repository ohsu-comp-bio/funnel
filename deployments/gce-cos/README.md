# Google Cloud Compute Deployment (Container-optimized)

This guide covers deploying Funnel to [Google Cloud Compute (GCE)][gce] using Docker containers and Google's [container-optimized OS][cos].

Previously, deploying Funnel to GCE required building and maintaining VM images. This approach replaces those VM images with Docker containers, which makes building and deploying VMs quicker and easier.

# Requirements

- [gcloud SDK][gcloud]

# Deployment

Run:
```
# Deploy a Funnel server
./make-server.sh

# Deploy a Funnel node
./make-node

# Create Funnel instance templates, which enables Funnel
# to automatically create nodes.
./make-node-templates.sh
```

[gce]: https://cloud.google.com/compute/
[cos]: https://cloud.google.com/container-optimized-os/docs/
[gcloud]: https://cloud.google.com/sdk/gcloud/
