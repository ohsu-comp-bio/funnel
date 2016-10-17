#!/usr/bin/env python

import unittest
import uuid
import time
import urllib
import json

from common_test_util import ServerTest, get_abspath

class TestTaskREST(ServerTest):

    def test_hello_world(self):

        task = {
            "name" : "TestEcho",
            "projectId" : "MyProject",
            "description" : "Simple Echo Command",
            "resources" : {},
            "docker" : [
                {
                    "imageName" : "ubuntu",
                    "cmd" : ["echo", "hello", "world"]
                }
            ]
        }

        u = urllib.urlopen("http://localhost:8000/v1/jobs", json.dumps(task))
        data = json.loads(u.read())
        job_id = data['value']
        
        while True:
            r = urllib.urlopen("http://localhost:8000/v1/jobs/%s" % (job_id))
            data = json.loads(r.read())
            if data["state"] not in ['Queued', "Running"]:
                break
            time.sleep(1)

        assert 'logs' in data
        assert data['logs'][0]['stdout'] == "hello world\n"



