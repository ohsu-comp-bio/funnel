#!/usr/bin/env python

import unittest
import uuid
import time
import urllib
import json

from common_test_util import SimpleServerTest, get_abspath


class TestTaskREST(SimpleServerTest):

    def test_hello_world(self):

        task = {
            "name": "TestEcho",
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

        for i in range(10):
            r = urllib.urlopen("http://localhost:8000/v1/jobs/%s" % (job_id))
            data = json.loads(r.read())
            if data["state"] not in ['Queued', "Running"]:
                break
            time.sleep(1)

        assert 'logs' in data
        assert data['logs'][0]['stdout'] == "hello world\n"
