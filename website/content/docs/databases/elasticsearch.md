---
title: Elasticsearch
menu:
  main:
    parent: Databases
---

# Elasticsearch

Funnel supports storing tasks and scheduler data in Elasticsearch (v8).

Config:
```yaml
Database: elastic

Elastic:
  # Prefix to use for indexes
  IndexPrefix: "funnel"
  URL: http://localhost:9200
  # Optional. Username for HTTP Basic Authentication.
  Username:
  # Optional. Password for HTTP Basic Authentication.
  Password:
  # Optional. Endpoint for the Elastic Service (https://elastic.co/cloud).
  CloudID:
  # Optional. Base64-encoded token for authorization; if set, overrides username/password and service token.
  APIKey:
  # Optional. Service token for authorization; if set, overrides username/password.
  ServiceToken:
```
