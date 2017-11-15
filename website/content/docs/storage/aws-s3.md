---
title: AWS S3
menu:
  main:
    parent: Storage
---

# AWS S3

Funnel supports using [AWS S3](https://aws.amazon.com/s3/) for file storage.

The S3 storage client is enabled by default, and will try to automatically
load credentials from the environment. Alternatively, you
may explicitly set the credentials in the worker config:

```
Worker:
  Storage:
    S3:
      Disabled: false
      AWS:
        # The maximum number of times that a request will be retried for failures.
        MaxRetries: 10
        # AWS Access key ID
        Key: ""
        # AWS Secret Access Key
        Secret: ""
```

### Example task
```
{
  "name": "Hello world",
  "inputs": [{
    "url": "s3://funnel-bucket/hello.txt",
    "path": "/inputs/hello.txt"
  }],
  "outputs": [{
    "url": "s3://funnel-bucket/output.txt",
    "path": "/outputs/hello-out.txt"
  }],
  "executors": [{
    "image": "alpine",
    "command": ["cat", "/inputs/hello.txt"],
    "stdout": "/outputs/hello-out.txt",
  }]
}
```

### Known issues

The S3 client is currently AWS specific and doesn't work on non-AWS S3 systems. See [issue #338](https://github.com/ohsu-comp-bio/funnel/issues/338).
