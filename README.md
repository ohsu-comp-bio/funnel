master: ![master-build-status](https://travis-ci.org/ohsu-comp-bio/funnel.svg?branch=master)


Task Execution Schema (TES)
===========================

[![Join the chat at https://gitter.im/ohsu-comp-bio/funnel](https://badges.gitter.im/ohsu-comp-bio/funnel.svg)](https://gitter.im/ohsu-comp-bio/funnel?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)


The Task Execution Schema proposal can be found at https://github.com/ga4gh/task-execution-schemas
The Protocol Buffer Based Schema can be found at https://github.com/ga4gh/task-execution-schemas/blob/master/proto/task_execution.proto
The swagger translation can be viewed at http://editor.swagger.io/#/?import=https://github.com/ga4gh/task-execution-schemas/raw/master/swagger/proto/task_execution.swagger.json

Example Task Message
```
{
    "name" : "TestMD5",
	"projectId" : "MyProject",
	"description" : "My Desc",
	"inputs" : [
		{
			"name" : "infile",
			"description" : "File to be MD5ed",
			"location" : "s3://my-bucket/input_file",
			"path" : "/tmp/test_file"
		}
	],
	"outputs" : [
		{
			"location" : "s3://my-bucket/output_file",
			"path" : "/tmp/test_out"
		}
	],
	"resources" : {
		"volumes" : [{
			"name" : "test_disk",
			"sizeGb" : 5,
			"mountPoint" : "/tmp"
		}]
	},
	"docker" : [
		{
			"imageName" : "ubuntu",
			"cmd" : ["md5sum", "/tmp/test_file"],
			"stdout" : "/tmp/test_out"
		}
	]
}
```

Example Task Message:
```
{
  "jobId" : "6E57CA6B-0BC7-44FB-BA2C-0CBFEC629C63",
  "metadata" : { Custom service metadata },
  "task" : {Task Message Above},
  "state" : "Running",
  "logs" : [
  	{ Job Log }
  ]
}
```

Example Job Log Message:
```
{
  "cmd" : ["md5sum", "/tmp/test_file"],
  "startTime" : "2016-09-18T23:08:27Z",
  "endTime" : "2016-09-18T23:38:00Z",
  "stdout": "f6913671da6018ff8afcb1517983ab24  test_file",
  "stderr": "",
  "exitCode" = 0
}
```

Example Task Conversation:

Get meta-data about service
```
GET /v1/jobs-service
```
Returns (from reference server)
{"storageConfig":{"baseDir":"/var/run/task-execution-server/storage","storageType":"sharedFile"}}


Post Job
```
POST /v1/jobs {JSON body message above}
Return:
{ "value" : "{job uuid}"}
```

Get Job Info
```
GET /v1/jobs/{job uuid}
```
Returns Job Body Example:
```
{
   "jobId" : "06b170b4-6ae8-4f11-7fc6-4417f1778b74",
   "logs" : [
      {
         "exitCode" : -1
      }
   ],
   "task" : {
      "projectId" : "test",
      "inputs" : [
         {
            "location" : "fs://README.md",
            "description" : "input",
            "path" : "/mnt/README.md",
            "name" : "input"
         }
      ],
      "name" : "funnel workflow",
      "taskId" : "06b170b4-6ae8-4f11-7fc6-4417f1778b74",
      "resources" : {
         "minimumRamGb" : 1,
         "minimumCpuCores" : 1,
         "volumes" : [
            {
               "sizeGb" : 10,
               "name" : "data",
               "mountPoint" : "/mnt"
            }
         ]
      },
      "outputs" : [
         {
            "location" : "fs://output/sha",
            "path" : "/mnt/sha",
            "name" : "stdout",
            "description" : "tool stdout"
         }
      ],
      "docker" : [
         {
            "imageName" : "bmeg/openssl",
            "workdir" : "/mnt/sha",
            "cmd" : [
               "openssl",
               "dgst",
               "-sha",
               "/mnt/README.md"
            ]
         }
      ],
      "description" : "CWL TES task"
   },
   "state" : "Error"
}
```

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
