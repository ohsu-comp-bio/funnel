---
title: Basic Auth

menu:
  main:
    parent: guides
    weight: 30
---

# Basic Auth

By default, a Funnel server allows open access to its API endpoints, but it can be configured to require basic password authentication. To enable this, include a password in your config file:

```yaml
Server:
  Password: abc123
```

Make sure to properly protect the configuration file so that it's not readable by everyone:
```bash
$ chmod 600 funnel.config.yml
```

To use the password, set the `FUNNEL_SERVER_PASSWORD` environment variable:
```bash
$ export FUNNEL_SERVER_PASSWORD=abc123
$ funnel task list
```
