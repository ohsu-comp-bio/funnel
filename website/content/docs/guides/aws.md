---
title: AWS Deployment

menu:
  main:
    parent: guides
    weight: 20
---

# Amazon Web Services

This guide covers deploying a Funnel server that leverages [DynamoDB][0] for storage
and [Batch][1] for task execution. You'll need to set up several resources 
using either the Funnel CLI or through the provided Amazon web UI.


## Create Required AWS Batch Resources

For Funnel to execute tasks on Batch, you must define a Compute Environment,
Job Queue and Job Definition. Amazon provides a quick start guide for these steps [here][2]. 

Additionally, you must define an IAM role for your Batch Job Definition. The role 
provides the job container with permissions to call the API actions that are 
specified in its associated policies on your behalf. For this configuration, 
these jobs need access to S3 and DynamoDB. 


Note, we reccommend creating the Job Definition with Funnel via `funnel aws batch create-job-definition`. 
Funnel expects the JobDefinition to start a `worker` process with a specific configuration. Only 
advanced users should consider making any substantial changes to this JobDefinition. 

### Create Resources With AWS

* Create Compute Environment - [link][3]
* Create Job Queue - [link][4]
* Define an EC2ContainerTaskRole with policies for managing access to S3 and DynamoDB - [link][5]
* Create a Job Definition - [link][6]


### Create Resources With Funnel

```
$ funnel aws batch create-all-resources

Create a compute environment, job queue and job definition in a specified region

Usage:
  funnel aws batch create-all-resources [flags]

Flags:
      --ComputEnv.InstanceTypes strings      The instances types that may be launched. You can also choose optimal to pick instance types on the fly that match the demand of your job queues. (default [optimal])
      --ComputEnv.MaxVCPUs int               The maximum number of EC2 vCPUs that an environment can reach. (default 256)
      --ComputEnv.MinVCPUs int               The minimum number of EC2 vCPUs that an environment should maintain. (default 0)
      --ComputEnv.SecurityGroupIds strings   The EC2 security groups that are associated with instances launched in the compute environment. If none are specified all security groups will be used.
      --ComputEnv.Subnets strings            The VPC subnets into which the compute resources are launched. If none are specified all subnets will be used.
      --ComputeEnv.Name string               The name of the compute environment. (default "funnel-compute-environment")
      --JobDef.Image string                  The docker image used to start a container. (default "docker.io/ohsucompbio/funnel:latest")
      --JobDef.JobRoleArn string             The Amazon Resource Name (ARN) of the IAM role that the container can assume for AWS permissions. A role will be created if not provided.
      --JobDef.MemoryMiB int                 The hard limit (in MiB) of memory to present to the container. (default 128)
      --JobDef.Name string                   The name of the job definition. (default "funnel-job-def")
      --JobDef.VCPUs int                     The number of vCPUs reserved for the container. (default 1)
      --JobQueue.Name string                 The name of the job queue. (default "funnel-job-queue")
      --JobQueue.Priority int                The priority of the job queue. Priority is determined in descending order. (default 1)
      --config string                        Funnel configuration file
  -h, --help                                 help for create-resources
      --region string                        Region in which to create the Batch resources (default "us-west-2")
```


## Configuring the Funnel Server

Since the tasks and logs are stored in DynamoDB the Funnel server can be turned 
on and off without data loss. 


Start the server:

```
funnel server run --config /path/to/config.yaml
```


Example configuration:

```YAML
Server:
  Database: "dynamodb"
  Databases:
    Dynamodb:
      TableBasename: "funnel"
      Region: "us-west-2"
      Credentials:
        Key: ""
        Secret: ""

Backend: "batch"
Backends:
  Batch:
    JobDef:
      Name: "funnel-job-def"
      Image: "docker.io/ohsu-comp-bio/funnel:latest"
      DefaultMemory: 128
      DefaultVcpus: 1
      JobRoleArn: "<PLACEHOLDER>"
    JobQueue: "<PLACEHOLDER>"
    Region: "us-west-2"
    Credentials:
      Key: ""
      Secret: ""
            
Worker:
  TaskReader: "dynamodb"
  TaskReaders:
    Dynamodb:
      TableBasename: "funnel"
      Region: "us-west-2"
      Credentials:
        Key: ""
        Secret: ""
  ActiveEventWriters:
    - "log"
    - "dynamodb"
  EventWriters:
    Dynamodb:
      TableBasename: "funnel"
      Region: "us-west-2"
      Credentials:
        Key: ""
        Secret: ""
  Storage:
    S3:
      Credentials:
        Key: ""
        Secret: ""
```

[0]: http://docs.aws.amazon.com/amazondynamodb/latest/developerguide/Introduction.html
[1]: http://docs.aws.amazon.com/batch/latest/userguide/what-is-batch.html
[2]: http://docs.aws.amazon.com/batch/latest/userguide/Batch_GetStarted.html#first-run-step-2
[3]: https://us-west-2.console.aws.amazon.com/batch/home?region=us-west-2#/compute-environments/new
[4]: https://us-west-2.console.aws.amazon.com/batch/home?region=us-west-2#/queues/new
[5]: https://console.aws.amazon.com/iam/home?region=us-west-2#/roles$new?step=permissions&selectedService=EC2ContainerService&selectedUseCase=EC2ContainerTaskRole
[6]: https://us-west-2.console.aws.amazon.com/batch/home?region=us-west-2#/job-definitions/new
