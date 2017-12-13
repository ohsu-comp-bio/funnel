---
title: HTTP(s)
menu:
  main:
    parent: Storage
---

# HTTP

Funnel supports downloading files from public URLs via GET reqests. No authentication
mechanism is allowed. This backend can be used to fetch objects from cloud storage 
providers exposed using presigned URLs.

The HTTP storage client is enabled by default, but may be explicitly disabled in the 
worker config:

```
HTTPStorage:
  Disabled: false
  # Timeout for http(s) GET requests.
  # In nanoseconds.
  Timeout: 60000000000 # 60 seconds
```

### Example task
```
{
  "name": "Hello world",
  "inputs": [{
    "url": "http://fakedomain.com/hello.txt",
    "path": "/inputs/hello.txt"
  }],
  "executors": [{
    "image": "alpine",
    "command": ["cat", "/inputs/hello.txt"],
  }]
}
```
