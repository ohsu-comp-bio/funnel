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

Config:
```
Database: datastore

Datastore:
  Project: <google project ID>
```
