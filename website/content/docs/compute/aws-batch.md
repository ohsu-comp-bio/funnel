---
title: AWS Batch
menu:
  main:
    parent: Compute
    weight: 20
---


# Amazon Batch

This guide covers deploying a Funnel server that leverages [DynamoDB][0] for storage
and [Batch][1] for task execution. You'll need to set up several resources 
using either the Funnel CLI or through the provided Amazon web console.

### Create Required AWS Batch Resources

For Funnel to execute tasks on Batch, you must define a Compute Environment,
Job Queue and Job Definition. Additionally, you must define an IAM role for your 
Batch Job Definition. The role provides the job container with permissions to call 
the API actions that are specified in its associated policies on your behalf. For 
this configuration, these jobs need access to S3 and DynamoDB. 

Note, we recommend creating the Job Definition with Funnel by running: `funnel aws batch create-job-definition`. 
Funnel expects the JobDefinition to start a `worker` with a specific configuration. Only 
advanced users should consider making any substantial changes to this Job Definition. 

AWS Batch tasks, by default, launch the ECS Optimized AMI which includes 
an 8GB volume for the operating system and a 22GB volume for Docker image and metadata 
storage. The default Docker configuration allocates up to 10GB of this storage to 
each container instance. You can read more about this AMI at:

http://docs.aws.amazon.com/AmazonECS/latest/developerguide/ecs-optimized_AMI.html

The process for [creating a custom AMI for AWS Batch][7] is well-documented. Because 
AWS Batch has the same requirements for your AMI as Amazon ECS, use the default 
Amazon ECS-optimized Amazon Linux AMI as a base and change it to better suite 
your tasks.

### Create Resources With AWS

Amazon provides a quick start guide with more information [here][2]. 

* Create a Compute Environment - [link][3]
* Create a custom AMI - [link][7]
* Create a Job Queue - [link][4]
* Define an EC2ContainerTaskRole with policies for managing access to S3 and DynamoDB - [link][5]
* Create a Job Definition - [link][6]

### Create Resources With Funnel

Funnel provides a utility to create the resources you will need to get up and running.
You will need to specify the AWS region to create these resources in using the `--region` flag.

_Note:_ this command assumes your environment contains your AWS credentials. These 
can be configured with the `aws configure` command.

```
$ funnel aws batch create-all-resources

Create a compute environment, job queue and job definition in a specified region

Usage:
  funnel aws batch create-all-resources [flags]

Flags:
      --ComputeEnv.InstanceTypes strings     The instances types that may be launched. You can also choose optimal to pick instance types on the fly that match the demand of your job queues. (default [optimal])
      --ComputeEnv.MaxVCPUs int              The maximum number of EC2 vCPUs that an environment can reach. (default 256)
      --ComputeEnv.MinVCPUs int              The minimum number of EC2 vCPUs that an environment should maintain. (default 0)
      --ComputeEnv.SecurityGroupIds strings  The EC2 security groups that are associated with instances launched in the compute environment. If none are specified all security groups will be used.
      --ComputeEnv.Subnets strings           The VPC subnets into which the compute resources are launched. If none are specified all subnets will be used.
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
      --region string                        Region in which to create the Batch resources.
```


## Configuring the Funnel Server

Since the tasks and logs are stored in DynamoDB the Funnel server can be turned 
on and off without data loss. 


Start the server:

```
funnel server run --config /path/to/config.yaml
```

Below is an example of the configuration you would need for the server had you 
run `funnel aws batch create-all-resources --region us-west-2`. Note that the `Key`
and `Secret` fields are left blank in the configuration of the componenets. This is because 
Funnel will, by default, try to will try to automatically load credentials from the environment. 
Alternatively, you may explicitly set the credentials in the config.

```YAML
Database: "dynamodb"
Compute: "aws-batch"
EventWriters:
  - "log"
  - "dynamodb"

Dynamodb:
  TableBasename: "funnel"
  Region: "us-west-2"
  Key: ""
  Secret: ""

Batch:
  JobDefinition: "funnel-job-def"
  JobQueue: "funnel-job-queue" 
  Region: "us-west-2"
  Key: ""
  Secret: ""
          
Worker:
  TaskReader: "dynamodb"

AmazonS3:
  Key: ""
  Secret: ""
```

### Known issues

The `Task.Resources.DiskGb` field does not have any effect. See [issue 317](https://github.com/ohsu-comp-bio/funnel/issues/317).

[0]: http://docs.aws.amazon.com/amazondynamodb/latest/developerguide/Introduction.html
[1]: http://docs.aws.amazon.com/batch/latest/userguide/what-is-batch.html
[2]: http://docs.aws.amazon.com/batch/latest/userguide/Batch_GetStarted.html#first-run-step-2
[3]: https://us-west-2.console.aws.amazon.com/batch/home?region=us-west-2#/compute-environments/new
[4]: https://us-west-2.console.aws.amazon.com/batch/home?region=us-west-2#/queues/new
[5]: https://console.aws.amazon.com/iam/home?region=us-west-2#/roles$new?step=permissions&selectedService=EC2ContainerService&selectedUseCase=EC2ContainerTaskRole
[6]: https://us-west-2.console.aws.amazon.com/batch/home?region=us-west-2#/job-definitions/new
[7]: http://docs.aws.amazon.com/batch/latest/userguide/create-batch-ami.html
