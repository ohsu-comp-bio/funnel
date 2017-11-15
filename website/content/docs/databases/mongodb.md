---
title: MongoDB
menu:
  main:
    parent: Databases
---

# MongoDB

Funnel supports storing tasks and scheduler data in MongoDB.

Config:
```
Server:
  Database: mongodb
  Databases:
    MongoDB:
      # Addresses for the seed servers.
      Addrs:
        - "localhost"
      # Database name used within MongoDB to store funnel data.
      Database: "funnel"
      Username: ""
      Password: ""
```

### Writing events from the worker

The worker can be configured to write events directly to Mongo, which avoids unnecessary RPC traffic to the Funnel server.
```
Worker:
  ActiveEventWriters:
    - log
    - mongodb
  EventWriters:
    MongoDB:
      Addrs:
        - "localhost"
      Database: "funnel"
      Username: ""
      Password: ""
```

### Known issues

We have an unpleasant duplication of config between the Worker and Server blocks. Track this in [issue 339](https://github.com/ohsu-comp-bio/funnel/issues/339).
