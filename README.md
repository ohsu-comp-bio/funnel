master: ![master-build-status](https://travis-ci.org/bmeg/task-execution-server.svg?branch=master)

develop: ![develop-build-status](https://travis-ci.org/bmeg/task-execution-server.svg?branch=develop)

# task-execution-server

## Requirements
- [Protocol Buffers](https://github.com/google/protobuf) if making changes to the schema.


## Initial tool install
```
make depends
```


## Build project
```
make
```

## Start task server
```
./bin/tes-server
```

## Start worker
```
./bin/tes-worker
```

## Get info about task execution service
```
curl http://localhost:8000/v1/jobs-service
```

## Get Task Execution Server CWL runner
```
git clone https://github.com/bmeg/funnel.git
cd funnel/
virtualenv venv
. venv/bin/activate
pip install cwltool
pip install pyyaml
```

## Run Example workflow
```
python funnel/main.py --tes tes.yaml test/hashsplitter-workflow.cwl --input README.md

```

## Python examples

There are some example/helper scripts in the `examples/` directory, which might be useful during development. For example, to submit 10 tasks to TES which each sleep for 5 seconds, run:
```
python examples/submit-sleep-tasks.py --count 10 --sleep 5
```
