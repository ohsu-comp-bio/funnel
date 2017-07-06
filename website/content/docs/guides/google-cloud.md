---
title: Google Cloud

menu:
  main:
    parent: guides
    weight: 20
---

# Google Cloud Compute

This guide covers deploying a Funnel server and workers to [Google Cloud Compute (GCE)][1].
You'll need to create a Google Cloud project, and install the [gcloud][2] SDK.
Optionally, these commands can be done through the GCE web dashboard.


## Create a Funnel VM Image

A Funnel [image][3] provides the basic dependencies and configuration
needed by both Funnel servers and workers, and allows instances to start
more quickly and reliably.

Manually creating a GCE image can be tedious, but the Funnel [image installer][4]
makes it a little easier and automated.

```bash
#!/bin/bash

INSTALLER='https://github.com/ohsu-comp-bio/funnel/releases/download/dev/bundle.run'

# Create a VM, download the installer, and run it.
# The installer will install dependencies and create an image.
gcloud compute instances create funnel-image-builder                \
  --scopes compute-rw                                               \
  --metadata "startup-script-url=$INSTALLER,serial-port-enable=1"

# Follow the logs
gcloud compute instances tail-serial-port-output $NAME
```

This takes a few minutes. The installer needs to create a VM instance,
snapshot, disk, and then finally an image. Images will be created in the
"funnel" [image family][imgfam].


## Create a Server

Now that you have an image, it's fairly easy to create a server:

```bash
#!/bin/bash

gcloud compute instances create funnel-server \
  --scopes       'compute-rw,storage-rw'      \
  --zone         'us-west1-a'                 \
  --tags         'funnel,http-server'         \
  --image-family 'funnel'
```


<h2>Create a Worker <i class="optional">optional</i></h2>

```bash
#!/bin/bash

# RPC address of the Funnel server.
FUNNEL_SERVER="funnel-server:9090"

# Create a worker instance using the Funnel image.
gcloud compute instances create funnel-worker-1 \
  --scopes compute-rw,storage-rw                \
  --zone 'us-west1-a'                           \
  --tags funnel                                 \
  --image-family funnel                         \
  --machine-type 'n1-standard-16'               \
  --boot-disk-type 'pd-standard'                \
  --boot-disk-size '250GB'                      \
  --metadata "funnel-worker-serveraddress=$FUNNEL_SERVER"
```


<h2>Create Worker Templates <i class="optional">optional</i></h2>

Funnel includes a GCE autoscaler which can automatically start workers as needed.
Funnel uses [instances templates][8] to describe the types of workers available.

The script below creates three templates for workers of various sizes:
1 CPU, 2 CPU, and 4 CPU.

```bash
#!/bin/bash

# List of machine types to create templates for.
# https://cloud.google.com/compute/docs/machine-types
MACHINE_TYPES="
n1-standard-1
n1-standard-2
n1-standard-4
"

for mt in $MACHINE_TYPES; do
  gcloud compute instance-templates create "funnel-worker-$mt" \
    --scopes compute-rw,storage-rw \
    --tags funnel \
    --image-family funnel \
    --machine-type $mt \
    --boot-disk-type 'pd-standard' \
    --boot-disk-size '250GB'
done
```


# Custom Installations

This guide and the example code in Funnel is just the tip of the iceberg, really.


[1]: https://cloud.google.com/compute/
[2]: https://cloud.google.com/sdk/gcloud/
[3]: https://cloud.google.com/compute/docs/images
[4]: https://github.com/ohsu-comp-bio/funnel/tree/master/deployments/gce/bundle
[8]: https://cloud.google.com/compute/docs/instance-templates
[imgfam]: https://cloud.google.com/compute/docs/images#image_families
