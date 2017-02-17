# Getting Started

## Prerequisites

Funnel requires [Docker](https://docker.io) to be installed on the worker machines.

## Installation

Two options:

1. Download a binary release from the [releases](/releases) page.
2. Build the code. See the [development docs](./development.md).


## Configuration

Funnel can be used in many different configurations: local, private and public clouds, manual or automatic provisioning, object storage, etc.

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

See the [configuration docs](./configuration.md) for full detail on configuration.


## Local mode

This is the simplest setup. This runs both the server and workers locally. The server will handle starting worker processes, and the local filesystem will be used for storage.

Run `tes-server`

Try out a "Hello, world!" task with `python examples/hello-world.py`

Check out the web dashboard at `http://localhost:8000`


## Manual cluster mode

In this mode, the Funnel server runs on one machine, and you manually start worker processes on other machines.

To start the server, run `tes-server`

To start the worker, run `tes-worker --server-address <address-of-server>:9090`

The Funnel RPC API runs on port 9090, which is configurable.

## Auto cluster mode

In this mode, the Funnel server will automatically start worker processes on other machines. For example, Funnel can start a Google Clould VM with a worker process automatically, or integrate with an existing HTCondor cluster.

***Work in progress***


## Python examples

There are example python scripts in the [examples directory](../examples). 

For example, to submit 10 tasks which each sleep for 5 seconds, run:
```
python examples/submit-sleep-tasks.py --count 10 --sleep 5
```


## Next steps

Check out the [API docs](./apis.md) to learn about the available API endpoints and schemas available.

Check out the [Guides](./guides) section for more detail about specific parts of Funnel such as schedulers, storage backends, and more.

Dig into the code! Writing Funnel code is easy and fun. Check out the [development docs](./development.md).
