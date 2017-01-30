#!/usr/bin/env python

import unittest
import uuid
import time
import urllib
import json

from common_test_util import SimpleServerTest, get_abspath


class TestTaskREST(SimpleServerTest):

    def _submit_steps(self, *steps):
        docker = []
        for s in steps:
            docker.append({
                "imageName": "ubuntu",
                "cmd": s.split(' '),
                "stdout": "stdout",
            })

        return self.tes.submit({
            "name": "TestCase",
            "projectId": "Project ID",
            "description": "Test case.",
            "resources": {},
            "docker": docker,
        })

    def test_hello_world(self):
        '''Test a basic "Hello world" task and expected API result.'''
        job_id = self._submit_steps("echo hello world")
        data = self.tes.wait(job_id)
        assert 'logs' in data
        assert data['logs'][0]['stdout'] == "hello world\n"


    def test_job_log_length(self):
        '''
        The job logs list should only include entries for steps that have
        been started or completed, i.e. steps that have yet to be started
        won't show up in Job.Logs.
        '''
        job_id = self._submit_steps(
            "sleep 10",
            "echo end",
        )
        time.sleep(1)
        data = self.tes.get_job(job_id)
        assert len(data['logs']) == 1


    def test_mark_complete_bug(self):
        '''
        There was a bug + fix where the job was being marked complete after
        the first step completed, but the correct behavior is to mark the
        job complete after *all* steps have completed.
        '''
        job_id = self._submit_steps(
            "echo step 1",
            "sleep 5",
            "echo step 2",
        )
        while True:
            data = self.tes.get_job(job_id)
            if len(data['logs']) == 2:
                assert data['state'] == 'Running'
            elif len(data['logs']) == 3:
                break
            time.sleep(1)
