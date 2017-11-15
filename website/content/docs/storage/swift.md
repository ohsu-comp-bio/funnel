---
title: OpenStack Swift
menu:
  main:
    parent: Storage
---

# OpenStack Swift

Funnel supports using [OpenStack Swift][swift] for file storage.

The Swift client is NOT enabled by default, you must explicitly give the credentials
in the worker config:
```
Worker:
  Storage:
      Swift:
        UserName:
        Password:
        AuthURL:
        TenantName:
        TenantID:
        RegionName:
```

The config currently only supports OpenStack v2 auth. See [issue #336](https://github.com/ohsu-comp-bio/funnel/issues/336).

As always, if you set the password in this file, make sure you protect it appropriately. Alternatively, the Swift client
can pull credentials from these environment variables: https://godoc.org/github.com/ncw/swift#Connection.ApplyEnvironment  

Swift currently fails while uploading large objects. See [issue #257](https://github.com/ohsu-comp-bio/funnel/issues/257).


### Example task
```
{
  "name": "Hello world",
  "inputs": [{
    "url": "swift://funnel-bucket/hello.txt",
    "path": "/inputs/hello.txt"
  }],
  "outputs": [{
    "url": "swift://funnel-bucket/output.txt",
    "path": "/outputs/hello-out.txt"
  }],
  "executors": [{
    "image": "alpine",
    "command": ["cat", "/inputs/hello.txt"],
    "stdout": "/outputs/hello-out.txt",
  }]
}
```

[swift]: https://docs.openstack.org/swift/latest/
