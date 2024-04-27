---
title: Download 0.11.0
menu:
  main:
    weight: -2000
---

{{< download-links >}}

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
$ git clone https://github.com/ohsu-comp-bio/funnel.git
$ cd funnel
$ make
```

Funnel requires Go 1.21+. Check out the [development docs][dev] for more detail.

### Release History

See the [Releases](https://github.com/ohsu-comp-bio/funnel/releases)  page for release history.


[dev]: /docs/development/developers/
[docker]: https://docker.io
