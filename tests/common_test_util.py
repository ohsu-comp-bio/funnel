
import unittest
import os
import tempfile
from urlparse import urlparse
import logging
import time
import socket
import subprocess

WORK_DIR = os.path.join(os.path.dirname(os.path.dirname(__file__)), "test_work" )


class ServerTest(unittest.TestCase):

    def setUp(self):
        if not os.path.exists("./test_tmp"):
            os.mkdir("test_tmp")
        self.task_server = None
        f, db_path = tempfile.mkstemp(dir="./test_tmp", prefix="ga4hg_task_db.")
        os.close(f)
        cmd = ["./bin/ga4gh-taskserver", "-db", db_path]
        logging.info("Running %s" % (" ".join(cmd)))
        self.task_server = subprocess.Popen(cmd)
        time.sleep(3)
        
        self.task_worker = None
        cmd = ["./bin/ga4gh-worker"]
        logging.info("Running %s" % (" ".join(cmd)))
        self.task_worker = subprocess.Popen(cmd)

    def tearDown(self):
        if self.task_server is not None:
            self.task_server.kill()
            
        if self.task_worker is not None:
            self.task_worker.kill()
