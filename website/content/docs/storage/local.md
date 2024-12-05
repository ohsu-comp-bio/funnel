---
title: Local
menu:
  main:
    parent: Storage
    weight: -10
---

# Local

Funnel supports using the local filesystem for file storage.

Funnel limits which directories may be accessed, by default only allowing directories 
under the current working directory of the Funnel worker.

Config:
```yaml
LocalStorage:
  # Whitelist of local directory paths which Funnel is allowed to access.
  AllowedDirs:
    - ./
    - /path/to/allowed/dir
    - ...etc
```

### Example task

Files must be absolute paths in `file:///path/to/file.txt` URL form.

```
{
  "name": "Hello world",
  "inputs": [{
    "url": "file:///path/to/funnel-data/hello.txt",
    "path": "/inputs/hello.txt"
  }],
  "outputs": [{
    "url": "file:///path/to/funnel-data/output.txt",
    "path": "/outputs/hello-out.txt"
  }],
  "executors": [{
    "image": "alpine",
    "command": ["cat", "/inputs/hello.txt"],
    "stdout": "/outputs/hello-out.txt",
  }]
}
```

### File hard linking behavior

For efficiency, Funnel will attempt not to copy the input files, instead trying 
create a hard link to the source file. In some cases this isn't possible. For example, 
if the source file is on a network file system mount (e.g. NFS) but the Funnel worker's 
working directory is on the local scratch disk, a hard link would cross a file system 
boundary, which is not possible. In this case, Funnel will copy the file.

### File ownership behavior

One difficult area of files and Docker containers is file owner/group management. 
If a Docker container runs as root, it's likely that the file will end up being owned 
by root on the host system. In this case, some step (Funnel or another task) will 
likely fail to access it. This is a tricky problem with no good solution yet. 
See [issue 66](https://github.com/ohsu-comp-bio/funnel/issues/66).
