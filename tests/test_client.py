#!/usr/bin/env python

import unittest
import uuid
import time
import requests

from common_test_util import ServerTest

class TestTaskREST(ServerTest):

    def test_job_create(self):        

        task = {
            "name" : "TestTask",
            "projectId" : "MyProject",
            "description" : "My Desc",
            "inputParameters" : [
            
            ],
            "outputParameters" : [
            
            ],
            "resources" : {},
            "docker" : [
                {
                    "imageName" : "ubuntu",
                    "cmd" : "echo hello world"
                }
            ]
        }

        task_op = {
            'taskArgs' : {
                "inputs" : {}
            },
            'ephemeralTask' : task
        }

        #r = requests.post("http://localhost:8000/v1/tasks", json=payload)
        #print r.text

        r = requests.post("http://localhost:8000/v1/tasks:run", json=task_op)
        data = r.json()
        taskop_id = data['value']
        
        while True:
            r = requests.get("http://localhost:8000/v1/taskop/%s" % (taskop_id))
            print r.text
            data = r.json()
            if data["state"] not in ['Queued', "Running"]:
                break
            time.sleep(1)


