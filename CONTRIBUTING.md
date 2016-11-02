
DEVELOPMENT
-----------

This reference server is built in GO, serving the protocol described in 
https://github.com/ga4gh/task-execution-schemas (note, the repo is linked as a
git submodule, so make sure to use `git clone --recursive`)

The code structure is based on two binary programs

1) tes-server: Provide client API endpoint, web UI, manages work queue.
The client facing HTTP port (by default 8000) uses a GO web router to serve static HTML
elements and proxy API requests to a GRPC endpoint, using the GRPC gateway code 
(https://github.com/grpc-ecosystem/grpc-gateway)
The GRPC endpoint (usually bound to port 9090) is based on the auto-generated GO 
protoc GRPC server. It is bound to two services: the GA4GH Task Execution Service 
and a internal service protocol that is used by the workers to access the work queue.

2) tes-worker: A client that uses the internal service protocol to contact task 
scheduler (on port 9090) and requests jobs. It is then responsible for obtaining the 
required input files (thus it much have code and credentials to access object store),
run the docker container with the provided arguments, and then copy the result files 
out to the object store.

Code Structure
--------------

Main function of taskserver program (client interface and scheduler)
```
src/tes-server/
```

Main function of worker program
```
src/tes-worker/
```
 
Code related to the worker, include file mapper, local and swift file system clients, docker interfaces and worker thread manager
```
src/tes/worker/
```
 
The compiled copy of the Task Execution Schema protobuf
```
src/tes/ga4gh
```
 
BoltDB based implementation of the TES api as well as the scheduler API.
```
src/tes/server
```
 
The compiled copy of the scheduler API protobuf
```
src/tes/server/proto/
```
 
Python driven unit/conformance tests
```
tests
```
 
HTML and angular app for web view
```
share
```


Rebuilding Proto code
---------------------
First install protoc extentions (GRPC and GRPC gateway builder programs)
```
make depends
```
Rebuild auto-generated code
```
make proto_build
```


Running Tests
-------------
The integration testing is done using Python scripts to drive client API requests.
To run tests:
```
nosetests tests
```





Task Execution Schema (TES)
---------------------------

The Task Execution Schema proposal can be found at 
https://github.com/ga4gh/task-execution-schemas
The Protocol Buffer Based Schema can be found at 
https://github.com/ga4gh/task-execution-schemas/blob/master/proto/task_execution.proto

The swagger translation can be viewed at 
http://editor.swagger.io/#/?import=https://github.com/ga4gh/task-execution-schemas/raw/master/swagger/proto/task_execution.swagger.json

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
```
{"storageConfig":{"baseDir":"/var/run/task-execution-server/storage","storageType":"sharedFile"}}
```

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

