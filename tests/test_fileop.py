#!/usr/bin/env python

import unittest
import uuid
import time
import requests

from common_test_util import ServerTest, get_abspath

class TestFileOP(ServerTest):


    def test_file_mount(self):

        task = {
            "name" : "TestMD5",
            "projectId" : "MyProject",
            "description" : "My Desc",
            "inputs" : [
                {
                    "name" : "infile",
                    "description" : "File to be MD5ed",
                    "storage" : "test_data.1",
                    "path" : "/tmp/test_file"
                }
            ],
            "outputs" : [
                {
                    "storage" : "test_data.out",
                    "path" : "/tmp/test_out"
                }
            ],
            "resources" : {
                "disks" : [{
                    "name" : "test_disk",
                    "sizeGb" : 5,
                    "autoDelete" : True,
                    "readOnly" : True,
                    "mountPoint" : "/tmp"
                }]
            },
            "docker" : [
                {
                    "imageName" : "ubuntu",
                    "cmd" : "md5sum /tmp/test_file > /tmp/test_out"
                }
            ]
        }

        r = requests.post("http://localhost:8000/v1/jobs", json=task)
        data = r.json()
        print data
        job_id = data['value']

        for i in range(10):
            r = requests.get("http://localhost:8000/v1/jobs/%s" % (job_id))
            data = r.json()
            if data["state"] not in ['Queued', "Running"]:
                break
            time.sleep(1)
        print data

        #assert 'logs' in data
        #assert data['logs'][0]['stdout'] == "hello world\n"


