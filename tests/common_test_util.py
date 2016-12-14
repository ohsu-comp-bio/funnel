
import py_tes
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

S3_ACCESS_KEY="AKIAIOSFODNN7EXAMPLE"
S3_SECRET_KEY="wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
API_TOKEN="secret"

BUCKET_NAME="tes-test"

class S3ServerTest(unittest.TestCase):

    def setUp(self):
        if not os.path.exists("./test_tmp"):
            os.mkdir("test_tmp")
        self.task_server = None
        f, db_path = tempfile.mkstemp(dir="./test_tmp", prefix="tes_task_db.")
        os.close(f)
        self.storage_dir = db_path + ".storage"
        self.volume_dir = db_path + ".volumes"
        self.output_dir = db_path + ".output"
        os.mkdir(self.storage_dir)
        os.mkdir(self.volume_dir)
        os.mkdir(self.output_dir)
        
        self.dir_name = os.path.basename(db_path)
        
        #start s3 server
        cmd = [
            which("docker"),
            "run", "-p", "9000:9000",
            "--rm",
            "--name", "tes_minio_test",
            "-e", "MINIO_ACCESS_KEY=%s" % (S3_ACCESS_KEY),
            "-e", "MINIO_SECRET_KEY=%s" % (S3_SECRET_KEY),
            "-v", "%s:/export" % (self.storage_dir),
            "minio/minio", "server", "/export"
        ]
        logging.info("Running %s" % (" ".join(cmd)))        
        self.s3_server = subprocess.Popen(cmd)
        cmd = ["./bin/tes-server", "-db", db_path, "-s3", "http://localhost:9000"]
        logging.info("Running %s" % (" ".join(cmd)))
        self.task_server = subprocess.Popen(cmd)
        time.sleep(3)
        
        self.task_worker = None
        cmd = ["./bin/tes-worker"]
        logging.info("Running %s" % (" ".join(cmd)))
        self.task_worker = subprocess.Popen(cmd)
        
        tes = py_tes.TES("http://localhost:8000", s3_access_key=S3_ACCESS_KEY, s3_secret_key=S3_SECRET_KEY, secret=API_TOKEN)
        if not tes.bucket_exists(BUCKET_NAME):
            tes.make_bucket(BUCKET_NAME)

    def tearDown(self):
        if self.s3_server is not None:
            self.s3_server.kill()
            cmd = ["docker", "kill", "tes_minio_test"]
            logging.info("Running %s" % (" ".join(cmd)))
            
            cmd = ["docker", "rm", "-fv", "tes_minio_test"]
            logging.info("Running %s" % (" ".join(cmd)))
            
            subprocess.Popen(cmd).communicate()
        
        if self.task_server is not None:
            self.task_server.kill()
            
        if self.task_worker is not None:
            self.task_worker.kill()
    
    def get_storage_url(self, path):
        dstpath = "s3://%s/%s" % (BUCKET_NAME, os.path.join(self.dir_name, os.path.basename(path)))
        return dstpath
    
    def copy_to_storage( self, path):
        dstpath = "s3://%s/%s" % (BUCKET_NAME, os.path.join(self.dir_name, os.path.basename(path)))
        tes = py_tes.TES("http://localhost:8000", s3_access_key=S3_ACCESS_KEY, s3_secret_key=S3_SECRET_KEY, secret=API_TOKEN)
        print "uploading:", dstpath
        tes.upload_file(path, dstpath)
        return dstpath
    
    def get_from_storage(self, loc):
        tes = py_tes.TES("http://localhost:8000", s3_access_key=S3_ACCESS_KEY, s3_secret_key=S3_SECRET_KEY, secret=API_TOKEN)
        dst = os.path.join(self.output_dir, os.path.basename(loc))
        print "Downloading %s to %s" % (loc, dst)
        tes.download_file(dst, loc)
        return dst

    