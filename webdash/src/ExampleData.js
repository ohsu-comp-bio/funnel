const example_task = {
  "id": "bnj8hlnpbjg64189lu30",
  "state": "COMPLETE",
  "name": "sh -c 'echo starting; cat $file1 \u003e $file2; echo done'",
  "tags": {
    "tag-ONE": "TWO",
    "tag-THREE": "FOUR"
  },
  "volumes": ["/vol1", "/vol2"],
  "inputs": [
    {
      "name": "file1",
      "url": "file:Users/strucka/go/src/github.com/ohsu-comp-bio/funnel/README.md",
      "path": "/inputs/Users/strucka/go/src/github.com/ohsu-comp-bio/funnel/README.md"
    }
  ],
  "outputs": [
    {
      "name": "stdout-0",
      "url": "file:Users/strucka/go/src/github.com/ohsu-comp-bio/funnel/test.stdout",
      "path": "/outputs/stdout-0"
    },
    {
      "name": "file2",
      "url": "file:Users/strucka/go/src/github.com/ohsu-comp-bio/funnel/test_out",
      "path": "/outputs/Users/strucka/go/src/github.com/ohsu-comp-bio/funnel/test_out"
    }
  ],
  "resources": {
    "cpuCores": 2,
    "ramGb": 4,
    "diskGb": 10
  },
  "executors": [
    {
      "image": "ubuntu",
      "command": [
        "sh",
        "-c",
        "echo starting; cat $file1 \u003e $file2; echo done"
      ],
      "stdout": "/outputs/stdout-0",
      "env": {
        "file1": "/inputs/Users/strucka/go/src/github.com/ohsu-comp-bio/funnel/README.md",
        "file2": "/outputs/Users/strucka/go/src/github.com/ohsu-comp-bio/funnel/test_out"
      }
    }
  ],
  "logs": [
    {
      "logs": [
        {
          "startTime": "2019-12-03T08:09:58.524782-08:00",
          "endTime": "2019-12-03T08:10:04.209567-08:00",
          "stdout": "starting\ndone\n"
        }
      ],
      "metadata": {
        "hostname": "BICB230"
      },
      "startTime": "2019-12-03T08:09:58.516832-08:00",
      "endTime": "2019-12-03T08:10:04.216273-08:00",
      "outputs": [
        {
          "url": "file:Users/strucka/go/src/github.com/ohsu-comp-bio/funnel/test.stdout",
          "path": "/outputs/stdout-0",
          "sizeBytes": "14"
        },
        {
          "url": "file:Users/strucka/go/src/github.com/ohsu-comp-bio/funnel/test_out",
          "path": "/outputs/Users/strucka/go/src/github.com/ohsu-comp-bio/funnel/test_out",
          "sizeBytes": "1209"
        }
      ],
      "systemLogs": [
        "level='info' msg='Version' timestamp='2019-12-03T08:09:58.5126-08:00' task_attempt='0' executor_index='0' GitCommit='a630947d' GitBranch='master' GitUpstream='git@github.com:ohsu-comp-bio/funnel.git' BuildDate='2019-01-29T00:50:30Z' Version='0.9.0'",
        "level='info' msg='download started' timestamp='2019-12-03T08:09:58.521417-08:00' task_attempt='0' executor_index='0' url='file:Users/strucka/go/src/github.com/ohsu-comp-bio/funnel/README.md'",
        "level='info' msg='download finished' timestamp='2019-12-03T08:09:58.523199-08:00' task_attempt='0' executor_index='0' url='file:Users/strucka/go/src/github.com/ohsu-comp-bio/funnel/README.md' size='1209' etag=''",
        "level='info' msg='Running command' timestamp='2019-12-03T08:10:02.83215-08:00' task_attempt='0' executor_index='0' cmd='docker run -i --read-only --rm -e file1=/inputs/Users/strucka/go/src/github.com/ohsu-comp-bio/funnel/README.md -e file2=/outputs/Users/strucka/go/src/github.com/ohsu-comp-bio/funnel/test_out --name bnj8hlnpbjg64189lu30-0 -v /Users/strucka/go/src/github.com/ohsu-comp-bio/funnel/funnel-work-dir/bnj8hlnpbjg64189lu30/tmp:/tmp:rw -v /Users/strucka/go/src/github.com/ohsu-comp-bio/funnel/funnel-work-dir/bnj8hlnpbjg64189lu30/inputs/Users/strucka/go/src/github.com/ohsu-comp-bio/funnel/README.md:/inputs/Users/strucka/go/src/github.com/ohsu-comp-bio/funnel/README.md:ro -v /Users/strucka/go/src/github.com/ohsu-comp-bio/funnel/funnel-work-dir/bnj8hlnpbjg64189lu30/outputs:/outputs:rw ubuntu sh -c echo starting; cat $file1 \u003e $file2; echo done'",
        "level='info' msg='upload started' timestamp='2019-12-03T08:10:04.211287-08:00' task_attempt='0' executor_index='0' url='file:Users/strucka/go/src/github.com/ohsu-comp-bio/funnel/test_out'",
        "level='info' msg='upload finished' timestamp='2019-12-03T08:10:04.213672-08:00' task_attempt='0' executor_index='0' size='14' url='file:Users/strucka/go/src/github.com/ohsu-comp-bio/funnel/test.stdout' etag=''"
      ]
    }
  ],
  "creationTime": "2019-12-03T08:09:58.506338-08:00"
};

const example_node = {
  "id": "5bd9947c-4222-423c-6718-208c6525caa3",
  "resources": {
    "cpus": 4,
    "ramGb": 17.179869184,
    "diskGb": 672.442179584
  },
  "available": {
    "cpus": 4,
    "ramGb": 17.179869184,
    "diskGb": 672.442179584
  },
  "taskIds": [
    "bnnd427pbjgb7lgg06gg",
    "bnnd427pbjgb7lgg06gg-foo"
  ],
  "state": "ALIVE",
  "hostname": "BICB230",
  "lastPing": "1575930967253002000"
};

const example_task_list = [
  {"id": "1", "state": "COMPLETE"},
  {"id": "2", "state": "RUNNING"},
];

const example_node_list = [
  {
    "id": "366ab31a-8dec-4691-7a24-59be053a3d55",
    "resources": {
      "cpus": 4,
      "ramGb": 17.179869184,
      "diskGb": 669.49789696
    },
    "available": {
      "cpus": 4,
      "ramGb": 17.179869184,
      "diskGb": 669.49789696
    },
    "state": "DEAD",
    "hostname": "BICB230",
    "lastPing": "1575930783260904000"
  },
  {
    "id": "5bd9947c-4222-423c-6718-208c6525caa3",
    "resources": {
      "cpus": 4,
      "ramGb": 17.179869184,
      "diskGb": 672.442179584
    },
    "available": {
      "cpus": 4,
      "ramGb": 17.179869184,
      "diskGb": 672.442179584
    },
    "taskIds": [
      "bnnd427pbjgb7lgg06gg",
      "bnnd427pbjgb7lgg06gg-foo"
    ],
    "state": "ALIVE",
    "hostname": "BICB230",
    "lastPing": "1575930967253002000"
  }
];

const example_service_info = {
  "name": "Funnel",
  "doc": "git commit: a630947d\ngit branch: master\ngit upstream: git@github.com:ohsu-comp-bio/funnel.git\nbuild date: 2019-01-29T00:50:30Z\nversion: 0.9.0",
  "taskStateCounts": {
    "CANCELED": 0,
    "COMPLETE": 0,
    "EXECUTOR_ERROR": 0,
    "INITIALIZING": 0,
    "PAUSED": 0,
    "QUEUED": 0,
    "RUNNING": 0,
    "SYSTEM_ERROR": 0,
    "UNKNOWN": 0
  }
};

export { example_task, example_node, example_service_info, example_task_list, example_node_list };
