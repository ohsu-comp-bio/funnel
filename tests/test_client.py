#!/usr/bin/env python

import unittest
import uuid
import time
import requests

from common_test_util import ServerTest, get_abspath

class TestTaskREST(ServerTest):

    def test_hello_world(self):

        task = {
            "name" : "TestEcho",
            "projectId" : "MyProject",
            "description" : "Simple Echo Command",
            "inputs" : [
            ],
            "outputs" : [
            ],
            "resources" : {},
            "docker" : [
                {
                    "imageName" : "ubuntu",
                    "cmd" : "echo hello world"
                }
            ]
        }

        r = requests.post("http://localhost:8000/v1/jobs", json=task)
        data = r.json()
        job_id = data['value']
        
        while True:
            r = requests.get("http://localhost:8000/v1/jobs/%s" % (job_id))
            print r.text
            data = r.json()
            if data["state"] not in ['Queued', "Running"]:
                break
            time.sleep(1)

        assert 'logs' in data
        assert data['logs'][0]['stdout'] == "hello world\n"



