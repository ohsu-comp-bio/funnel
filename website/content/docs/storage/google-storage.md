---
title: Google Storage
menu:
  main:
    parent: Storage
---

# Google Storage

Funnel supports using [Google Storage][gs] (GS) for file storage.

The GS client is NOT enabled by default, you must enabled it in the config:
```
Worker:
  Storage:
      GS:
          # Automatically discover credentials from the environment.
        - FromEnv: true
          # Path to account credentials file.
          AccountFile:
```

In the near future, Google Storage will be enabled by default. See [issue #332](https://github.com/ohsu-comp-bio/funnel/issues/332).

### Example task
```
{
  "name": "Hello world",
  "inputs": [{
    "url": "gs://funnel-bucket/hello.txt",
    "path": "/inputs/hello.txt"
  }],
  "outputs": [{
    "url": "gs://funnel-bucket/output.txt",
    "path": "/outputs/hello-out.txt"
  }],
  "executors": [{
    "image": "alpine",
    "command": ["cat", "/inputs/hello.txt"],
    "stdout": "/outputs/hello-out.txt",
  }]
}
```

[gs]: https://cloud.google.com/storage/
