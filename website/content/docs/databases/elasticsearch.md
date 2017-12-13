---
title: Elasticsearch
menu:
  main:
    parent: Databases
---

# Elasticsearch

Funnel supports storing tasks and scheduler data in Elasticsearch.

Config:
```
Database: elastic
EventWriters:
  # log all events
  - log

Worker:
  # get the task directly from the database
  TaskReader: elastic

Elastic:
  # Prefix to use for indexes
  IndexPrefix: "funnel"
  URL: http://localhost:9200
```
