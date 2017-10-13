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

For Funnel to execute tasks on Batch, you must define a Compute Environment and 
Job Queue. Amazon provides a quick start guide for these steps [here][2]. 

Additionally, you must define an IAM role for your Batch Job Definition. The role 
provides the job container with permissions to call the API actions that are 
specified in its associated policies on your behalf. For this configuration, 
these jobs need access to S3 and DynamoDB. Since Funnel manages the creation of 
Job Definitons for Batch, this IAM role, along with some addtional options are 
provided in the Config for Funnel. 


### Create Resources With AWS

* Create Compute Environment - [link][3]
* Create Job Queue - [link][4]
* (OPTIONAL) Define custom policies for managing access to S3 and DynamoDB
* Define an EC2ContainerTaskRole - [link][5]


### Create Resources With Funnel

```
funnel aws batch create
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
