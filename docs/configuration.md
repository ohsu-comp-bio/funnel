# Configuration

***Work in progress***

Default configuration (generated)
```yaml
WorkDir: funnel-work-dir
DBPath: funnel-work-dir/funnel_task.db

# Run the 
HTTPPort: "8000"

# Run the RPC server on this port
RPCPort: "9090"

# The hostname of the Funnel server.
# Used by workers for communication.
HostName: localhost

LogLevel: debug

# Maximum per-task-step log size in bytes
MaxTaskLogSize: 10000

# Storage backend configuration
Storage: null
- Local
    # You must explicitly give Funnel access to local directories.
    AllowedDirs:
    - /home/ubuntu/funnel-storage
    - /tmp
    - /opt/foo/bar
- GS:
    AccountFile: /path/to/google-account-key-file.json

# Per-worker configuration
Worker:
  ServerAddress: localhost:9090
  LogLevel: debug
  LogPath: ""
  LogTailSize: 10000
  LogUpdateRate: 5000000000
  NewTaskPollRate: 5000000000
  # In nanoseconds (TODO fix this)
  StatusPollRate: 5000000000
  # Configure the worker to shutdown if it is idle.
  # -1 means never time out.
  # 0 means shut down immediately when there are no tasks.
  Timeout: -1
  WorkDir: funnel-work-dir

# Active scheduler backend
Scheduler: local
```
