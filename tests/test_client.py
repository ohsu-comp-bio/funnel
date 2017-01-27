#!/usr/bin/env python

import time
import urllib2
import json

from common_test_util import SimpleServerTest


class TestTaskREST(SimpleServerTest):
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

    def test_hello_world(self):
        u = urllib2.urlopen("http://localhost:8000/v1/jobs",
                            json.dumps(self.task))
        data = json.loads(u.read())
        self.job_id = data['value']

        for i in range(10):
            r = urllib2.urlopen("http://localhost:8000/v1/jobs/%s" % (self.job_id))
            data = json.loads(r.read())
            if data["state"] not in ['Queued', "Running"]:
                break
            time.sleep(1)

        assert data["state"] == "Complete"
        assert 'logs' in data
        assert data['logs'][0]['stdout'] == "hello world\n"

    def state_immutability(self):
        r = urllib2.urlopen("http://localhost:8000/v1/jobs/%s" % (self.job_id))
        old_data = json.loads(r.read())
        req = urllib2.Request("http://localhost:8000/v1/jobs/%s" % (job_id))
        req.get_method = lambda: 'DELETE'
        response = urllib2.urlopen(req)
        new_data = json.loads(response.read())
        assert new_data["state"] == old_data["state"]
