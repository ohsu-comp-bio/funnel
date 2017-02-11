import argparse
import json
import urllib

parser = argparse.ArgumentParser()
parser.add_argument("--count", type=int, default=1)
parser.add_argument("--sleep", type=int, default=10)

args = parser.parse_args()

task = {
    "name": "Date loop",
    "projectId": "TES development",
    "description": "Print the date every second, for-ev-or.",
    "inputs": [],
    "outputs": [],
    "resources": {},
    "docker": [
        {
            "imageName": "ubuntu",
            "cmd": ["bash", "-c", "while true; do echo a; sleep 10; done"],
            "stdout": "/tmp/date_loop_out",
            "stderr": "/tmp/date_loop_err",
        }
    ]
}

for x in range(args.count):
    u = urllib.urlopen("http://localhost:8000/v1/jobs", json.dumps(task))
