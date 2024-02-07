---
title: Nextflow
menu:
  main:
    parent: Interop
---

# Nextflow

[Nextflow](https://nextflow.io/) is a workflow engine with a [rich ecosystem]() of pipelines centered around biological analysis.

> Nextflow enables scalable and reproducible scientific workflows using software containers. It allows the adaptation of pipelines written in the most common scripting languages.

> Its fluent DSL simplifies the implementation and the deployment of complex parallel and reactive workflows on clouds and clusters. 

Since Nextflow [includes support](https://www.nextflow.io/docs/latest/executor.html#ga4gh-tes) for the TES API, it can be used in conjunction with Funnel to run tasks or to interact with a common TES endpoint.  

## Getting Started

To set up Nextflow to use Funnel as the TES executor, run the following:

```
funnel server run
```

## Additional Resources

- [Nextflow Homepage](https://nextflow.io/)

- [Nextflow Documentation](https://www.nextflow.io/docs)

- [Nextflow's TES Support](https://www.nextflow.io/docs/latest/executor.html#ga4gh-tes)

- [nf-core](https://nf-co.re/)
  > A community effort to collect a curated set of analysis pipelines built using Nextflow. 

- [nf-canary](https://github.com/seqeralabs/nf-canary)
  > A minimal Nextflow workflow for testing infrastructure. 
