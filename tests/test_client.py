#!/usr/bin/env python

import unittest
import uuid
import time
import requests

from common_test_util import ServerTest, get_abspath

class TestTaskREST(ServerTest):

    def test_job_run(self):

        task = {
            "name" : "TestEcho",
            "projectId" : "MyProject",
            "description" : "Simple Echo Command",
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

        assert 'logs' in data
        assert data['logs'][0]['stdout'] == "hello world\n"


    def test_file_mount(self):

        task = {
            "name" : "TestMD5",
            "projectId" : "MyProject",
            "description" : "My Desc",
            "inputParameters" : [
                {
                    "name" : "infile",
                    "description" : "File to be MD5ed",
                    "localCopy" : {
                        "disk" : "test_disk",
                        "path" : "/tmp/test_file"
                    }
                }
            ],
            "outputParameters" : [

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
                    "cmd" : "md5sum /tmp/test_file"
                }
            ]
        }

        task_op = {
            'taskArgs' : {
                "inputs" : {
                    "infile" : get_abspath("test_data.1")
                }
            },
            'ephemeralTask' : task
        }
        r = requests.post("http://localhost:8000/v1/tasks:run", json=task_op)
        print "ON point"
        print r
        data = r.json()
        print data
        taskop_id = data['value']

        for i in range(10):
            r = requests.get("http://localhost:8000/v1/taskop/%s" % (taskop_id))
            print r.text
            data = r.json()
            if data["state"] not in ['Queued', "Running"]:
                break
            time.sleep(1)

        #assert 'logs' in data
        #assert data['logs'][0]['stdout'] == "hello world\n"


