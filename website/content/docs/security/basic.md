---
title: Basic Auth
menu:
  main:
    parent: Security
    weight: 10
---
# Basic Auth

By default, a Funnel server allows open access to its API endpoints, but it 
can be configured to require basic password authentication. To enable this, 
include users and passwords in your config file:

```yaml
Server:
  BasicAuth:
    - User: funnel
      Password: abc123
```

If you are using BoltDB or Badger, the Funnel worker communicates to the server via gRPC
so you will also need to configure the RPC client. 

```yaml
RPCClient:
  User: funnel
  Password: abc123
```

Make sure to properly protect the configuration file so that it's not readable 
by everyone:

```bash
$ chmod 600 funnel.config.yml
```

To use the password, set the `FUNNEL_SERVER_USER` and `FUNNEL_SERVER_PASSWORD` environment variables:
```bash
$ export FUNNEL_SERVER_USER=funnel
$ export FUNNEL_SERVER_PASSWORD=abc123
$ funnel task list
```
