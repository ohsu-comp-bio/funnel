---
title: Quickstart

menu:
  main:
    weight: -70
---

# TODO or maybe Getting Started

Here's a simple `funnel.config.yml` file for getting started:
```yaml
# Describes the storage that Funnel has access to.
Storage:
- Local:
    # You need to explicitly give Funnel access to local directories.
    AllowedDirs:
    - /home/buchanae/funnel-files
    - /tmp
```
