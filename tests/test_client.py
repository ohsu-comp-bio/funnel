#!/usr/bin/env python

import docker
import json
import os
import time
import urllib2

from common_test_util import SimpleServerTest


TESTS_DIR = os.path.dirname(__file__)
blocking_util = os.path.join(TESTS_DIR, "blocking_util.py")


class TestTaskREST(SimpleServerTest):

    def _submit_steps(self, *steps):
        docker = []
        for s in steps:
            docker.append({
                "imageName": "tes_tests",
                "cmd": s.split(' '),
                "stdout": "stdout",
                "ports": [{
                    "host": 5000,
                    "container": 5000
                }]
            })
        task = {
            "name": "TestCase",
            "projectId": "Project ID",
            "description": "Test case.",
            "inputs": [
                {
                    "name": "blocking_util",
                    "description": "testing util",
                    "location": "file://" + blocking_util,
                    "class": "File",
                    "path": "/tmp/blocking_util.py"
                }
            ],
            "resources": {
                "volumes": [{
                    "name": "test_disk",
                    "sizeGb": 5,
                    "mountPoint": "/tmp"
                }]
            },
            "docker": docker
        }
        print json.dumps(task)
        return self.tes.submit(task)

    def test_hello_world(self):
        '''Test a basic "Hello world" task and expected API result.'''
        job_id = self._submit_steps("echo hello world")
        data = self.tes.wait(job_id)
        assert 'logs' in data
        assert data['logs'][0]['stdout'] == "hello world\n"

    def test_state_immutability(self):
        job_id = self._submit_steps("echo hello world")
        data = self.tes.wait(job_id)
        self.tes.delete_job(job_id)
        new_data = self.tes.get_job(job_id)
        assert new_data["state"] == data["state"]

    def test_cancel(self):
        job_id = self._submit_steps("sleep 100")
        time.sleep(5)
        self.tes.delete_job(job_id)
        time.sleep(5)
        data = self.tes.get_job(job_id)
        assert data["state"] == "Canceled"
        time.sleep(10)
        dclient = docker.from_env()
        self.assertRaises(Exception, dclient.containers.get, job_id + "-0")

    def test_job_log_length(self):
        '''
        The job logs list should only include entries for steps that have
        been started or completed, i.e. steps that have yet to be started
        won't show up in Job.Logs.
        '''
        job_id = self._submit_steps(
            "python3 /tmp/blocking_util.py",
            "echo end"
        )
        time.sleep(5)
        data = self.tes.get_job(job_id)
        assert len(data['logs']) == 1
        # send shutdown signal to blocking_util
        urllib2.urlopen("http://127.0.0.1:5000/shutdown")
        time.sleep(5)
        # 2nd step should start now
        data = self.tes.get_job(job_id)
        assert len(data['logs']) == 2

    def test_mark_complete_bug(self):
        '''
        There was a bug + fix where the job was being marked complete after
        the first step completed, but the correct behavior is to mark the
        job complete after *all* steps have completed.
        '''
        job_id = self._submit_steps(
            "echo step 1",
            "python3 /tmp/blocking_util.py",
            "echo step 2",
        )
        while True:
            data = self.tes.get_job(job_id)
            if 'logs' in data:
                if len(data['logs']) == 2:
                    assert data['state'] == 'Running'
                elif len(data['logs']) == 3:
                    break
            time.sleep(1)
