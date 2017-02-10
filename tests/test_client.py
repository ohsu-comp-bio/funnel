#!/usr/bin/env python

import docker
import json
import os
import time

from common_test_util import SimpleServerTest


TESTS_DIR = os.path.dirname(__file__)


class TestTaskREST(SimpleServerTest):

    def _submit_steps(self, *steps):
        docker = []
        for s in steps:
            docker.append({
                "imageName": "tes-wait",
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
            "resources": {},
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
        dclient = docker.from_env()
        job_id = self._submit_steps("tes-wait step 1", "tes-wait step 2")
        self.wait("step 1", timeout=20)
        # ensure the container was created
        self.wait_for_container(job_id + "-0")
        self.tes.delete_job(job_id)
        # ensure docker stops the container within 20 seconds
        self.wait_for_container_stop(job_id + "-0", timeout=20)
        # make sure the first container was stopped
        self.assertRaises(Exception, dclient.containers.get, job_id + "-0")
        # make sure the second container was never started
        self.assertRaises(Exception, dclient.containers.get, job_id + "-1")
        data = self.tes.get_job(job_id)
        assert data["state"] == "Canceled"

    def test_job_log_length(self):
        '''
        The job logs list should only include entries for steps that have
        been started or completed, i.e. steps that have yet to be started
        won't show up in Job.Logs.
        '''
        job_id = self._submit_steps(
            "tes-wait step 1",
            "echo done"
        )
        self.wait("step 1")
        data = self.tes.get_job(job_id)
        assert len(data['logs']) == 1
        self.resume()

    def test_mark_complete_bug(self):
        '''
        There was a bug + fix where the job was being marked complete after
        the first step completed, but the correct behavior is to mark the
        job complete after *all* steps have completed.
        '''
        job_id = self._submit_steps(
            "echo step 1",
            "tes-wait step 2",
            "echo step 2",
        )
        self.wait("step 2")
        data = self.tes.get_job(job_id)
        assert 'logs' in data
        assert data['state'] == 'Running'
        self.resume()
