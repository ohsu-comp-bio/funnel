---
title: Embedded
menu:
  main:
    parent: Databases
    weight: -10
---

# Embedded

By default, Funnel uses an embedded database named [BoltDB][bolt] to store task 
and scheduler data. This is great for development and a simple server without 
external dependencies, but it doesn't scale well to larger clusters.

Available config:
```
Database: boltdb

BoltDB:
  # Path to database file
  Path: ./funnel-work-dir/funnel.db
```

[bolt]: https://github.com/boltdb/bolt
