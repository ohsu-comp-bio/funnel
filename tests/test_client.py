from __future__ import print_function

import docker
import json
import os
import time
import shlex

from common_test_util import SimpleServerTest


TESTS_DIR = os.path.dirname(__file__)


class TestClient(SimpleServerTest):

    def dumps(self, d):
        return json.dumps(d, indent=2, sort_keys=True)

    def _submit_steps(self, *steps):
        executor = []
        for s in steps:
            executor.append({
                "image_name": "tes-wait",
                "cmd": shlex.split(s),
                "stdout": "stdout",
                "ports": [{
                    "host": 5000,
                    "container": 5000
                }]
            })
        task = {
            "name": "TestCase",
            "project": "Project ID",
            "description": "Test case.",
            "resources": {},
            "executors": executor
        }
        print('submitted task', self.dumps(task))
        return self.tes.submit(task)

    def test_hello_world(self):
        '''
        Test a basic "Hello world" task and expected API result.
        '''
        task_id = self._submit_steps("echo hello world")
        data = self.tes.wait(task_id)
        print(self.dumps(data))
        assert 'logs' in data
        assert len(data['logs']) == 1
        assert 'logs' in data['logs'][0]
        assert data['logs'][0]['logs'][0]['stdout'] == "hello world\n"

    def test_single_char_log(self):
        '''
        Test that the streaming logs pick up a single character.

        This ensures that the streaming works even when a small
        amount of logs are written.
        '''
        task_id = self._submit_steps("bash -c 'echo a; tes-wait step 1'")
        self.wait("step 1")
        time.sleep(0.1)
        data = self.tes.get_task(task_id)
        print(self.dumps(data))
        assert 'logs' in data
        assert 'logs' in data['logs'][0]
        assert data['logs'][0]['logs'][0]['stdout'] == "a\n"
        self.resume()

    def test_port_log(self):
        '''
        Ensure that ports are logged and returned correctly.
        '''
        task_id = self._submit_steps("tes-wait step 1")
        self.wait("step 1")
        time.sleep(0.1)
        data = self.tes.get_task(task_id)
        print(self.dumps(data))
        assert 'logs' in data
        assert 'logs' in data['logs'][0]
        assert data['logs'][0]['logs'][0]['ports'][0]['host'] == 5000
        self.resume()

    def test_state_immutability(self):
        task_id = self._submit_steps("echo hello world")
        data = self.tes.wait(task_id)
        self.tes.cancel_task(task_id)
        new_data = self.tes.get_task(task_id)
        assert new_data["state"] == data["state"]

    def test_cancel(self):
        dclient = docker.from_env()
        task_id = self._submit_steps("tes-wait step 1", "tes-wait step 2")
        self.wait("step 1", timeout=20)
        # ensure the container was created
        self.wait_for_container(task_id + "-0")
        self.tes.cancel_task(task_id)
        # ensure docker stops the container within 20 seconds
        self.wait_for_container_stop(task_id + "-0", timeout=20)
        # make sure the first container was stopped
        self.assertRaises(Exception, dclient.containers.get, task_id + "-0")
        # make sure the second container was never started
        self.assertRaises(Exception, dclient.containers.get, task_id + "-1")
        data = self.tes.get_task(task_id)
        print(self.dumps(data))
        assert data["state"] == "CANCELED"

    def test_executor_log_length(self):
        '''
        The task executor logs list should only include entries for steps that have
        been started or completed, i.e. steps that have yet to be started
        won't show up in Task.Logs[0].Logs
        '''
        task_id = self._submit_steps(
            "tes-wait step 1",
            "echo done"
        )
        self.wait("step 1")
        data = self.tes.get_task(task_id)
        print(self.dumps(data))
        assert len(data['logs']) == 1
        assert len(data['logs'][0]['logs']) == 1
        self.resume()

    def test_mark_complete_bug(self):
        '''
        There was a bug + fix where the task was being marked complete after
        the first step completed, but the correct behavior is to mark the
        task complete after *all* steps have completed.
        '''
        task_id = self._submit_steps(
            "echo step 1",
            "tes-wait step 2",
            "echo step 2",
        )
        self.wait("step 2")
        data = self.tes.get_task(task_id)
        print(self.dumps(data))
        assert data['state'] == 'RUNNING'
        self.resume()
