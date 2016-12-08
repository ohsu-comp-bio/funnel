
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

class SimpleServerTest(unittest.TestCase):

    def setUp(self):
        if not os.path.exists("./test_tmp"):
            os.mkdir("test_tmp")
        self.task_server = None
        f, db_path = tempfile.mkstemp(dir="./test_tmp", prefix="tes_task_db.")
        os.close(f)
        self.storage_dir = db_path + ".storage"
        self.volume_dir = db_path + ".volumes"
        os.mkdir(self.storage_dir)
        os.mkdir(self.volume_dir)
        
        cmd = ["./bin/tes-server", "-db", db_path, "-storage", self.storage_dir]
        logging.info("Running %s" % (" ".join(cmd)))
        self.task_server = subprocess.Popen(cmd)
        time.sleep(3)
        
        self.task_worker = None
        cmd = ["./bin/tes-worker", "-volumes", self.volume_dir]
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
    
    def get_from_storage(self, loc):
        dst = os.path.join( self.storage_dir, loc )
        return dst


def which(file):
    for path in os.environ["PATH"].split(":"):
        p = os.path.join(path, file)
        if os.path.exists(p):
            return p


class S3ServerTest(unittest.TestCase):

    def setUp(self):
        if not os.path.exists("./test_tmp"):
            os.mkdir("test_tmp")
        self.task_server = None
        f, db_path = tempfile.mkstemp(dir="./test_tmp", prefix="tes_task_db.")
        os.close(f)
        self.storage_dir = db_path + ".storage"
        self.volume_dir = db_path + ".volumes"
        os.mkdir(self.storage_dir)
        os.mkdir(self.volume_dir)
        
        #start s3 server
        cmd = [
            which("docker"),
            "run", "-p", "9000:9000",
            "-v", "%s:/export" % (self.storage_dir),
            "minio/minio", "server", "/export"
        ]
        cmd = ["./bin/tes-server", "-db", db_path, "-s3", "http://localhost:9000"]
        logging.info("Running %s" % (" ".join(cmd)))
        self.task_server = subprocess.Popen(cmd)
        time.sleep(3)
        
        self.task_worker = None
        cmd = ["./bin/tes-worker", "-volumes", self.volume_dir]
        logging.info("Running %s" % (" ".join(cmd)))
        self.task_worker = subprocess.Popen(cmd)

    def tearDown(self):
        if self.task_server is not None:
            self.task_server.kill()
            
        if self.task_worker is not None:
            self.task_worker.kill()
    