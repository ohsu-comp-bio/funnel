import argparse
import json
import urllib

parser = argparse.ArgumentParser()
parser.add_argument("--count", type=int, default=1)
parser.add_argument("--sleep", type=int, default=10)

args = parser.parse_args()

task = {
    "name": "TestMD5",
    "projectId": "MyProject",
    "description": "My Desc",
    "inputs": [
        {
            "name": "infile",
            "description": "File to be MD5ed",
            "location": "file:///tmp/test_file",
            "class": "File",
            "path": "/tmp/test_file"
        }
    ],
    "outputs": [
        {
            "location": "file:///tmp/test_out_file",
            "class": "File",
            "path": "/tmp/test_out"
        }
    ],
    "resources": {
        "volumes": [{
            "name": "test_disk",
            "sizeGb": 5,
            "mountPoint": "/tmp"
        }]
    },
    "docker": [
        {
            "imageName": "ubuntu",
            "cmd": ["md5sum", "/tmp/test_file"],
            "stdout": "/tmp/test_out",
            "stderr": "/tmp/test_err",
        }
    ]
}

for x in range(args.count):
    u = urllib.urlopen("http://localhost:8000/v1/jobs", json.dumps(task))
