# Google Cloud Compute

This guide covers deploying a Funnel server and workers to [Google Cloud Compute (GCE)][1].
You'll need a Google Cloud project, and the [gcloud][2] SDK.


## Create a Funnel image

A Funnel [image][3] provides the basic dependencies and configuration
needed by both Funnel servers and workers, and allows instances to start
more quickly and reliably.

Creating a GCE image can be tedious, but the Funnel [image installer][4]
makes it a little easier and automated.

```bash
INSTALLER='https://github.com/ohsu-comp-bio/funnel/releases/download/dev/bundle.run'

# Create a VM, download the installer, and run it.
# The installer will install dependencies and create an image.
gcloud compute instances create $NAME \
  --scopes compute-rw \
  --metadata "startup-script-url=$INSTALLER,serial-port-enable=1"

# Follow the logs
gcloud compute instances tail-serial-port-output $NAME
```

This takes a couple minutes. The installer needs to create a VM instance,
snapshot, disk, and then finally an image. Images will be created in the
"funnel" [image family][].


## Create a Funnel server

Now that you have an image, it's fairly easy to create a server:

```bash
gcloud compute instances create funnel-server \
  --scopes       'compute-rw,storage-rw'      \
  --zone         'us-west1-a'                 \
  --tags         'funnel,http-server'         \
  --image-family 'funnel'
```


## (Optional) Create a Funnel worker

```bash
# Address of the Funnel server with RPC port
FUNNEL_SERVER="funnel-server:9090"

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


## (Optional) Create Funnel instance templates

In order to automatically start workers, Funnel needs worker [instances templates][8]
that describe what kinds of workers it can start. 

```bash
FUNNEL_SERVER='funnel-server:9090'
MACHINE_TYPES="
n1-standard-1
n1-standard-8
n1-standard-16
"

for mt in $MACHINE_TYPES; do
  NAME="funnel-worker-$mt"
  gcloud compute instance-templates create $NAME \
    --scopes compute-rw,storage-rw \
    --zone 'us-west1-a' \
    --tags funnel \
    --image-family funnel \
    --machine-type $mt \
    --boot-disk-type 'pd-standard' \
    --boot-disk-size '250GB' \
    --metadata "funnel-worker-serveraddress=$SERVER"
done
```


# Custom Installations

This guide and the example code in Funnel is just the tip of the iceberg, really.



[1]: https://cloud.google.com/compute/
[2]: https://cloud.google.com/sdk/gcloud/
[3]: https://cloud.google.com/compute/docs/images
[4]: ../../deployments/gce/bundle
[8]: https://cloud.google.com/compute/docs/instance-templates
