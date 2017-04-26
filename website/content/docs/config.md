---
title: Config

menu:
  main:
    weight: -60
---

# Configuration

Usage:
```shell
$ funnel --config funnel.config.yaml [server | worker | ... ]
```

Below is a YAML config file containing all the possible options and default values.  
(Note: the config is still a mess, but we're working on improving it)

```YAML
# Path to the database file.
DBPath: ./funnel-work-dir/funnel.db
# Port used for HTTP communication and the web dashboard.
HTTPPort: 8000
# Port used for RPC communication.
RPCPort: 9090
# Working files created during processing will be written in this directory.
WorkDir: ./funnel-work-dir
# Hostname of the Funnel server. Mainly, this is used by auto-scalers
# to configure workers.
HostName: localhost
# Logging levels: debug, info, error
LogLevel: debug
# Write logs to this path. If empty, logs are written to stderr.
LogPath: ""
# Limit the size of task executor logs (stdout/err), in bytes.
MaxExecutorLogSize: 10000 # 10 KB
# How often to run a scheduler iteration.
# In nanoseconds.
ScheduleRate: 1000000000 # 1 second
# How many tasks to schedule in one iteration.
ScheduleChunk: 10
# How long to wait between updates before marking a worker dead.
# In nanoseconds.
WorkerPingTimeout: 60000000000 # 1 minute
# How long to wait for a worker to start, before marking the worker dead.
# In nanoseconds.
WorkerInitTimeout: 300000000000 # 5 minutes
# Include a "Cache-Control: no-store" HTTP header in Get/List responses
# to prevent caching by intermediary services.
DisableHTTPCache: true

# File storage systems.
# This is a list of structs (TODO that's confusing, restructure this config)
Storage:
  # Local file system.
  - Local:
      # Whitelist of local directory paths which Funnel is allowed to access.
      AllowedDirs:
        - ./funnel-work-dir/storage
    #S3:
    #  Endpoint:
    #  Key:
    #  Secret:

     # Google Cloud Storage
     #GS:
      # Path to account credentials file.
      # Optional. If possible, credentials will be automatically discovered
      # from the environment.
     # AccountFile:
      # Automatically discover credentials from the environment.
     # FromEnv: true

# The name of the active scheduler backend
# Available backends: local, condor, gce, openstack
Scheduler: local

# Scheduler backend config
Backends:

  GCE:
    # Path to account credentials file.
    # If possible, this will be automatically discovered.
    AccountFile: ""
    # Google Cloud project ID.
    # If possible, this will be automatically discovered.
    Project: ""
    # Google Cloud zone.
    # If possible, this will be automatically discovered.
    Zone: ""
    Weights:
      # Prefer workers that start up quickly.
      # Workers that are already online have instant startup time.
      PreferQuickStartup: 1.0
    # How long to cache GCE API results (machine list, templates, etc)
    # before refreshing.
    CacheTTL: 60000000000 # 1 minute

Worker:
  # If empty, a worker ID will be automatically generated.
  ID: ""
  # RPC address of the Funnel server
  ServerAddress: localhost:9090
  # Files created during processing will be written in this directory.
  WorkDir: ./funnel-work-dir
  # If the worker has been idle for longer than the timeout, it will shut down.
  # -1 means there is no timeout. 0 means timeout immediately after the first task.
  Timeout: -1
  # Maximum task log (stdout/err) size, in bytes.
  LogTailSize: 10000 # 10 KB

  # File storage systems.
  #
  # This is usually set automatically by the scheduler, but you might need this
  # if you're starting a worker manually.
  Storage:
    # Local file system.
    - Local:
        # Whitelist of local directory paths which Funnel is allowed to access.
        AllowedDirs:
          - ./funnel-work-dir/storage
      #S3:
      #  Endpoint:
      #  Key:
      #  Secret:

       # Google Cloud Storage
       #GS:
        # Path to account credentials file.
        # Optional. If possible, credentials will be automatically discovered
        # from the environment.
       # AccountFile:
        # Automatically discover credentials from the environment.
       # FromEnv: true
  # Logging levels: debug, info, error
  LogLevel: debug
  # Write logs to this path. If empty, logs are written to stderr.
  LogPath: ""
  # Override available resources.
  Resources:
    # CPUs available.
    # Cpus: 0

    # RAM available, in GB.
    # Ram: 0.0

    # Disk space available, in GB.
    Disk: 100.0
  # For low-level tuning.
  # RPC timeout for update/sync call.
  # In nanoseconds.
  UpdateTimeout: 1000000000 # 1 second
  # For low-level tuning.
  # How often to sync with the Funnel server.
  # In nanoseconds.
  UpdateRate: 5000000000 # 5 seconds
  # For low-level tuning.
  # How often to send task log updates to the Funnel server.
  # In nanoseconds.
  LogUpdateRate: 5000000000 # 5 seconds
```
