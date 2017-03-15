import argparse
import json
import urllib

parser = argparse.ArgumentParser()
parser.add_argument("--count", type=int, default=1)

args = parser.parse_args()

task = {
    "name": "4 CPU sleep",
    "projectId": "MyProject",
    "description": "My Desc",
    "inputs": [
    ],
    "outputs": [
    ],
    "resources": {
        "minimumCpuCores": 4,
        "volumes": [{
            "name": "test_disk",
            "sizeGb": 5,
            "mountPoint": "/tmp"
        }]
    },
    "docker": [
        {
            "imageName": "ubuntu",
            "cmd": ["sleep", "30"],
            "stdout": "/tmp/test_out",
            "stderr": "/tmp/test_err",
        }
    ]
}

for x in range(args.count):
    u = urllib.urlopen("http://localhost:8000/v1/jobs", json.dumps(task))
