import json
from pprint import pprint
import time
import urllib

task = {
    "name": "Hello world",
    "projectId": "MyProject",
    "description": "Simple Echo Command",
    "resources": {},
    "docker": [
        {
            "imageName": "ubuntu",
            "cmd": ["echo", "hello", "world"],
            "stdout": "stdout",
        }
    ]
}

u = urllib.urlopen("http://localhost:8000/v1/jobs", json.dumps(task))
data = json.loads(u.read())
job_id = data['value']

while True:
    r = urllib.urlopen("http://localhost:8000/v1/jobs/%s" % (job_id))
    data = json.loads(r.read())
    if data["state"] not in ['Queued', 'Initializing', "Running"]:
        break
    time.sleep(1)

pprint(data)
assert 'logs' in data
assert data['logs'][0]['stdout'] == "hello world\n"
