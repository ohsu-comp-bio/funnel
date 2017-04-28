---
title: Install
menu:
  main:
    weight: -80
---

### Prerequisites

Funnel requires [Docker][docker].

### Download

Currently, we only have development releases:

- [Linux (64-bit)][linux-64-bin]
- [Linux (32-bit)][linux-32-bin]
- [macOS (64-bit)][mac-64-bin]

Windows is not supported yet.

<h3>Build the code <i class="optional">optional</i></h3>

In short, this will get you started:
```shell
$ go get github.com/ohsu-comp-bio/funnel
```
Check out the [development docs][dev] for more detail.




[linux-64-bin]: https://github.com/ohsu-comp-bio/funnel/releases/download/dev/funnel-linux-amd64.tar.gz
[linux-32-bin]: https://github.com/ohsu-comp-bio/funnel/releases/download/dev/funnel-linux-386.tar.gz
[mac-64-bin]: https://github.com/ohsu-comp-bio/funnel/releases/download/dev/funnel-darwin-amd64.tar.gz
[dev]: /docs/development/
[docker]: https://docker.io
