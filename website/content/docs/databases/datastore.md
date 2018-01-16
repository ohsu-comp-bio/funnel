---
title: Datastore
menu:
  main:
    parent: Databases
---

# Google Cloud Datastore

Funnel supports storing tasks (but not scheduler data) in Google Cloud Datastore.

This implementation currently doesn't work with Appengine, since Appengine places
special requirements on the context of requests and requires a separate library.

Two entity types are used, "Task" and "TaskPart" (for larger pieces of task content,
such as stdout/err logs).

Funnel will, by default, try to will try to automatically load credentials from the environment. Alternatively, you may explicitly set the credentials in the config.

Config:
```
Database: datastore

Datastore:
  Project: ""
  # Path to account credentials file.
  # Optional. If possible, credentials will be automatically discovered
  # from the environment.
  CredentialsFile: ""
```
