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
Database: mongodb
EventWriters:
  # log all events
  - log

Worker:
  # get the task directly from the database
  TaskReader: mongodb

MongoDB:
  # Addresses for the seed servers.
  Addrs:
    - "localhost"
  # Database name used within MongoDB to store funnel data.
  Database: "funnel"
  Username: ""
  Password: ""
```
