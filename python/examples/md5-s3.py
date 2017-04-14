import argparse
import json
import urllib.request as request

import jwt

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
            "location": "s3://tmp/test_out_file",
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

signing_key = 'secret'
for x in range(args.count):
    token = jwt.encode({
        'S3_ACCESS_KEY_ID': 'O8CEYH06QVNB36R2G4Z',
        'S3_ACCESS_SECRET': '0uxDczNgSJb5UuaFduf69lFcmC3IHNLcc2WY8acB',
    }, signing_key, algorithm='HS256').decode('utf-8')
    data = json.dumps(task).encode('utf-8')
    req = request.Request("http://localhost:8000/v1/jobs", data, {
        'authorization': 'JWT %s' % (token),
    })
    print('token', token)
    request.urlopen(req)
