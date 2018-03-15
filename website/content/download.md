---
title: Download
menu:
  main:
    weight: -2000
---

### Download

- [linux <small>[funnel-linux-amd64-0.6.0.tar.gz]</small>][linux-64-bin]
- [mac <small>[funnel-darwin-amd64-0.6.0.tar.gz]</small>][mac-64-bin]
- <small>Windows is not supported (yet), sorry!</small>

[linux-64-bin]: https://github.com/ohsu-comp-bio/funnel/releases/download/0.6.0/funnel-linux-amd64-0.6.0.tar.gz
[mac-64-bin]: https://github.com/ohsu-comp-bio/funnel/releases/download/0.6.0/funnel-darwin-amd64-0.6.0.tar.gz

Funnel is a single binary.  
Funnel requires [Docker][docker].  
Funnel is beta quality. APIs might break, bugs exist, data might be lost.  

### Homebrew

```
brew tap ohsu-comp-bio/formula
brew install funnel
```

<h3>Build the lastest development version <i class="optional">optional</i></h3>

In order to build the latest code, run:
```shell
$ go get github.com/ohsu-comp-bio/funnel
```

Funnel requires Go 1.8+. Check out the [development docs][dev] for more detail.

### Release History

See the [Releases](https://github.com/ohsu-comp-bio/funnel/releases)  page for release history.


[dev]: /docs/development/
[docker]: https://docker.io
