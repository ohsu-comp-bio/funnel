---
title: MongoDB
menu:
  main:
    parent: Databases
---

# MongoDB

Funnel supports storing tasks and scheduler data in MongoDB.

Config:
```yaml
Database: mongodb

MongoDB:
  # Addresses for the seed servers.
  Addrs:
    - "localhost"
  # Database name used within MongoDB to store funnel data.
  Database: "funnel"
  Username: ""
  Password: ""
```
