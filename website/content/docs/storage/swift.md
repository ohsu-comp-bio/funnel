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

The config currently only supports OpenStack v2 auth, which unfortunately requires the password to be saved in the config file. Make sure to protect this file appropriately. See [issue #336](https://github.com/ohsu-comp-bio/funnel/issues/336).

In the future, Swift will be enabled by default, with automatic detection of credentials from the environment. See [issue #253](https://github.com/ohsu-comp-bio/funnel/issues/253).

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
