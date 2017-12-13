---
title: DynamoDB
menu:
  main:
    parent: Databases
---

# DynamoDB

Funnel supports storing task data in DynamoDB. Storing scheduler data is not supported currently, so using the node scheduler with DynamoDB won't work. Using AWS Batch for compute scheduling may be a better option.
Funnel will, by default, try to will try to automatically load credentials from the environment. Alternatively, you may explicitly set the credentials in the config.

Available Config:
```
Database: dynamodb
EventWriters:
  # log all events
  - log

Worker:
  # get the task directly from the database
  TaskReader: dynamodb

DynamoDB:
  # Basename to use for dynamodb tables
  TableBasename: "funnel"
  # AWS region
  Region: "us-west-2"
  # AWS Access key ID
  Key: ""
  # AWS Secret Access Key
  Secret: ""
```

### Known issues

Dynamo does not store scheduler data. See [issue 340](https://github.com/ohsu-comp-bio/funnel/issues/340).
