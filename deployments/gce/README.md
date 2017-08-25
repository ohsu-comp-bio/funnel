# Google Cloud Compute Deployment

DEPRECATED: This guide uses VM images for deployment, but using Docker containers is much quicker and easier. See deployments/gce-cos.

This guide covers deploying a Funnel server and nodes to [Google Cloud Compute (GCE)][1].
You'll need to create a Google Cloud project, and install the [gcloud][2] SDK.


## Create a Funnel VM Image

A Funnel [image][3] provides the basic dependencies and configuration
needed by both Funnel servers and nodes, and allows instances to start
more quickly and reliably.

Manually creating a GCE image can be tedious, but the Funnel [image installer][4]
makes it a little easier and automated.

Run:
```
bash ./make-image.sh
```

This takes a few minutes. The installer needs to create a VM instance,
snapshot, disk, and then finally an image. Images will be created in the
"funnel" [image family][imgfam].


## Create a Server

Now that you have an image, it's fairly easy to create a server.

Run:
```
bash ./make-server.sh
```


<h2>Create a Node <i class="optional">optional</i></h2>

Create a node;
```bash
bash ./make-node.sh
```


<h2>Create Node Templates <i class="optional">optional</i></h2>

Funnel includes a GCE autoscaler which can automatically start nodes as needed.
Funnel uses [instances templates][8] to describe the types of nodes available.

The script below creates three templates for nodes of various sizes:
1 CPU, 2 CPU, and 4 CPU.

```bash
bash ./make-node-templates.sh
```


[1]: https://cloud.google.com/compute/
[2]: https://cloud.google.com/sdk/gcloud/
[3]: https://cloud.google.com/compute/docs/images
[4]: https://github.com/ohsu-comp-bio/funnel/tree/master/deployments/gce/bundle
[8]: https://cloud.google.com/compute/docs/instance-templates
[imgfam]: https://cloud.google.com/compute/docs/images#image_families
