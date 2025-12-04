---
title: Postgres
menu:
  main:
    parent: Databases
---

# Postgres

> [!WARNING]
>
> Postgres support is in early development.
>
> Production deployments are recommended to use the alternative storage backends available, such as MongoDB, ElasticSearch, DynamoDB, etc.

## Configuration

```yaml
Database: postgres

Postgres:
  Addrs:
    - "localhost"
  Database: "funnel"
  Username: ""
  Password: ""
```
