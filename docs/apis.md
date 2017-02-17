api

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
