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
Server:
  Database: elastic
  Databases:
    Elastic:
      # Prefix to use for indexes
      IndexPrefix: "funnel"
      URL: http://localhost:9200
```

### Writing events from the worker

The worker can be configured to write events directly to Elasticsearch, which avoids unnecessary RPC traffic to the Funnel server.
```
Worker:
  ActiveEventWriters:
    - log
    - elastic
  EventWriters:
    Elastic:
      # Prefix to use for indexes
      IndexPrefix: "funnel"
      URL: http://localhost:9200
```

### Known issues

We have an unpleasant duplication of config between the Worker and Server blocks. Track this in [issue 339](https://github.com/ohsu-comp-bio/funnel/issues/339).
