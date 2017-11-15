---
title: DynamoDB
menu:
  main:
    parent: Databases
---

# DynamoDB

Funnel supports storing task data in DynamoDB. Storing scheduler data is not supported currently, so using the node scheduler with DynamoDB won't work. Using AWS Batch for compute scheduling may be a better option.

Available Config:
```
Server:
  Database: dynamodb
  Databases:
    DynamoDB:
      # Basename to use for dynamodb tables
      TableBasename: "funnel"
      # AWS region
      Region: ""
      Credentials:
        # AWS Access key ID
        Key: ""
        # AWS Secret Access Key
        Secret: ""
```

### Worker config

Using DynamoDB with AWS Batch requires that the worker be configured to connect to the database:
```
Worker:
  ActiveEventWriters:
    - log
    - dynamodb
  EventWriters:
    DynamoDB:
      # Basename to use for dynamodb tables
      TableBasename: "funnel"
      # AWS region
      Region: ""
      Credentials:
        # AWS Access key ID
        Key: ""
        # AWS Secret Access Key
        Secret: ""

  TaskReader: dynamodb
  TaskReaders:
    DynamoDB:
      # Basename to use for dynamodb tables
      TableBasename: "funnel"
      # AWS region
      Region: ""
      Credentials:
        # AWS Access key ID
        Key: ""
        # AWS Secret Access Key
        Secret: ""
```

### Known issues

We have an unpleasant duplication of config between the Worker and Server blocks. Track this in [issue 339](https://github.com/ohsu-comp-bio/funnel/issues/339).

Dynamo does not store scheduler data. See [issue 340](https://github.com/ohsu-comp-bio/funnel/issues/340).
