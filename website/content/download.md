---
title: Download
menu:
  main:
    weight: -2000
---

### Download

- [linux <small>[funnel-linux-amd64-0.5.0.tar.gz]</small>][linux-64-bin]
- [mac <small>[funnel-darwin-amd64-0.5.0.tar.gz]</small>][mac-64-bin]
- <small>Windows is not supported (yet), sorry!</small>

[linux-64-bin]: https://github.com/ohsu-comp-bio/funnel/releases/download/0.5.0/funnel-linux-amd64-0.5.0.tar.gz
[mac-64-bin]: https://github.com/ohsu-comp-bio/funnel/releases/download/0.5.0/funnel-darwin-amd64-0.5.0.tar.gz

Funnel is a single binary.  
Funnel requires [Docker][docker].  
Funnel is beta quality. APIs might break, bugs exist, data might be lost.  

### Homebrew

```
brew tap ohsu-comp-bio/formula
brew install funnel
```

<h3>Install the lastest development version <i class="optional">optional</i></h3>

In order to build the latest code, run:
```shell
$ go get github.com/ohsu-comp-bio/funnel
```

Funnel requires Go 1.8+. Check out the [development docs][dev] for more detail.

### Release History

#### 0.5.0

Date: Jan 9, 2018
Tag: https://github.com/ohsu-comp-bio/funnel/releases/tag/0.5.0  
Changes: https://github.com/ohsu-comp-bio/funnel/compare/0.4.1...0.5.0  
Milestone: https://github.com/ohsu-comp-bio/funnel/milestone/3?closed=1  

- **Breaking changes**
  - Configuration
      - The configuration file structure has been refactored to simplify, remove large duplicated blocks and deep nesting. Most structures live at the root level now.
      - Added some basic config validation to catch misspelled or unknown fields, a common source of issues.
      - Added most config values to the available CLI flags.
- Failure tolerance
  - Added retries with exponential backoff and jitter to database and RPC clients.
- Databases
  - Added Google Cloud Datastore database backend.
- Events
  - Added “task created” event type.
  - Added full support for writing events to Kafka.
  - Added storage of system log events.
- Web dashboard
  - Display per-task system logs.
  - Tweaked display of large text fields such as stdout/err and input content.
- Storage
  - Ensure all file handles are closed consistently.
  - Added retries to Swift storage.
  - Produce warning on empty directory download.
  - Better defaults for chunk size in Swift backend.
  - Added generic S3 (i.e. not Amazon S3) storage backend/client, based on the Minio client library.
  - Also added the ability to configure/enable multiple, separate S3 backends simultaneously.
  - Added HTTP storage backend, which currently supports read-only operations (write/put is not supported).
- Removed the autoscaler code. This code was getting old and outdated, and nobody seemed to be using it. A fresh version will be rewritten in the future.
- Lots of other bugfixes.


#### 0.4.1

Date: Nov 16, 2017  
Tag: https://github.com/ohsu-comp-bio/funnel/releases/tag/0.4.1  
Changes: https://github.com/ohsu-comp-bio/funnel/compare/0.4.0...0.4.1  
Milestone: https://github.com/ohsu-comp-bio/funnel/milestone/4?closed=1  

Notes:

- Fix for regression in Elasticsearch scheduler database backend.

#### 0.4.0

Date: Nov 15, 2017  
Tag: https://github.com/ohsu-comp-bio/funnel/releases/tag/0.4.0  
Changes: https://github.com/ohsu-comp-bio/funnel/compare/0.3.0...0.4.0  
Milestone: https://github.com/ohsu-comp-bio/funnel/milestone/2?closed=1  

Notes:

- Upgrade task API to TES v0.3
  - **This is a database schema change. Tasks from previous versions are not compatible.**
- Added MongoDB support.
- Added Kafka event stream support.
- Bug fix for OutputFileLog path.
- Rewrote TES HTTP client to match gRPC client.
- Web dashboard cleanup and fixes: style cleanup, pagination fixes.
- Scheduler performance and scalability improvements, concurrency fixes.
- Pagination support in CLI and clients.
- Executor stdout/err tail performance and scalability improvements.
- Remove network inspection and logging of host IP and docker ports.
  Remove port requests. Part of TES v0.3.
- cmd/task/create bugfix, couldn't read task input file.
- Website and docs rewrite.

#### 0.3.0

Date: Nov 1, 2017  
Tag: https://github.com/ohsu-comp-bio/funnel/releases/tag/0.3.0  
Changes: https://github.com/ohsu-comp-bio/funnel/compare/0.2.0...0.3.0  
Milestone: https://github.com/ohsu-comp-bio/funnel/milestone/1?closed=1  

Notes:

- Added AWS DynamoDB, Batch, and S3 support.
- Added Elasticsearch database.
- Added Swift object storage client.
- Added task events schema.
- Web dashboard sorting, auto refresh, page size, and lots of other improvements.
- Run `docker pull` before `docker run` on each task to ensure the local images
  are up-to-date.
- Improved scalability and performance of scheduler and database.
- CLI tweaks and fixes
  - Use `FUNNEL_SERVER` environment variable to set funnel server URL.
  - `funnel run` now wraps command in a shell by default.
    Use `funnel run --exec` to bypass this.
  - Added `funnel version`
  - `funnel wait` moved to `funnel task wait`

#### 0.2.0

Date: Jul 18, 2017  
Tag: https://github.com/ohsu-comp-bio/funnel/releases/tag/0.2.0  
Changes: https://github.com/ohsu-comp-bio/funnel/compare/0.1.0...0.2.0  

Notes:

- Added/improved Google Cloud Compute autoscaling support and deployment scripts.
- Implemented pagination for ListTasks endpoint, in server only.
- Added basic auth. support

##### 0.1.0

Released: Jun 5, 2017

Tag: https://github.com/ohsu-comp-bio/funnel/releases/tag/0.1.0


[dev]: /docs/development/
[docker]: https://docker.io
