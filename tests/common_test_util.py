
import unittest
import os
import tempfile
from urlparse import urlparse
import logging
import shutil
import time
import socket
import subprocess

WORK_DIR = os.path.join(os.path.dirname(os.path.dirname(__file__)), "test_work" )

def get_abspath(path):
    return os.path.join(os.path.dirname(__file__), path)

class ServerTest(unittest.TestCase):

    def setUp(self):
        if not os.path.exists("./test_tmp"):
            os.mkdir("test_tmp")
        self.task_server = None
        f, db_path = tempfile.mkstemp(dir="./test_tmp", prefix="ga4hg_task_db.")
        os.close(f)
        self.storage_dir = db_path + ".storage"
        self.volume_dir = db_path + ".volumes"
        os.mkdir(self.storage_dir)
        os.mkdir(self.volume_dir)
        cmd = ["./bin/ga4gh-taskserver", "-db", db_path]
        logging.info("Running %s" % (" ".join(cmd)))
        self.task_server = subprocess.Popen(cmd)
        time.sleep(3)
        
        self.task_worker = None
        cmd = ["./bin/ga4gh-worker", "-volumes", self.volume_dir, "-storage", self.storage_dir]
        logging.info("Running %s" % (" ".join(cmd)))
        self.task_worker = subprocess.Popen(cmd)

    def tearDown(self):
        if self.task_server is not None:
            self.task_server.kill()
            
        if self.task_worker is not None:
            self.task_worker.kill()
    
    def copy_to_storage( self, path):
        dst = os.path.join( self.storage_dir, os.path.basename(path) )
        shutil.copy(path, dst)
        return os.path.basename(path)

