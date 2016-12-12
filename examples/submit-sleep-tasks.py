import argparse
import json
from pprint import pprint
import time
import urllib

parser = argparse.ArgumentParser()
parser.add_argument("--count", type=int, default=1)
parser.add_argument("--sleep", type=int, default=10)

args = parser.parse_args()

task = {
    "name" : "TestEcho",
    "projectId" : "MyProject",
    "description" : "Simple Echo Command",
    "resources" : {},
    "docker" : [
	{
	    "imageName" : "ubuntu",
	    "cmd" : ["sleep", args.sleep]
	}
    ]
}

for x in range(args.count):
    u = urllib.urlopen("http://localhost:8000/v1/jobs", json.dumps(task))
