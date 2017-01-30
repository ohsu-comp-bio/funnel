
import unittest
import os
import tempfile
from urlparse import urlparse
import logging
import shutil
import time
import socket
import subprocess
import logging
import signal

import yaml

import py_tes


S3_ENDPOINT = "localhost:9000"
S3_ACCESS_KEY="AKIAIOSFODNN7EXAMPLE"
S3_SECRET_KEY="wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
BUCKET_NAME="tes-test"
WORK_DIR = os.path.join(os.path.dirname(os.path.dirname(__file__)), "test_work" )

def popen(*args, **kwargs):
    kwargs['preexec_fn'] = os.setsid
    return subprocess.Popen(*args, **kwargs)

def kill(p):
    try:
        os.killpg(os.getpgid(p.pid), signal.SIGTERM)
        p.wait()
    except OSError:
        pass

def get_abspath(path):
    return os.path.join(os.path.dirname(__file__), path)

def which(file):
    for path in os.environ["PATH"].split(":"):
        p = os.path.join(path, file)
        if os.path.exists(p):
            return p

def temp_config(config):
    configFile = tempfile.NamedTemporaryFile(delete=False)
    yaml.dump(config, configFile)
    return configFile


class SimpleServerTest(unittest.TestCase):

    def setUp(self):
        self.addCleanup(self.cleanup)
        if not os.path.exists("./test_tmp"):
            os.mkdir("test_tmp")
        self.task_server = None
        f, db_path = tempfile.mkstemp(dir="./test_tmp", prefix="tes_task_db.")
        os.close(f)
        self.storage_dir = os.path.abspath(db_path + ".storage")
        os.mkdir(self.storage_dir)

        # Build server config file (YAML)
        configFile = temp_config({
            "ServerAddress": "localhost:9090",
            "DBPath": db_path,
            "WorkDir": "test_tmp",
            "Storage": [{
                "local": {
                    "allowed_dirs": [self.storage_dir]
                }
            }]
        })
        
        # Start server
        cmd = ["./bin/tes-server", "-config", configFile.name]
        logging.info("Running %s" % (" ".join(cmd)))
        self.task_server = popen(cmd)
        time.sleep(3)
        self.tes = py_tes.TES("http://localhost:8000")
        
    # We're using this instead of tearDown because python doesn't call tearDown
    # if setUp fails. Since our setUp is complex, that means things don't get
    # properly cleaned up (e.g. processes are orphaned).
    def cleanup(self):
        if self.task_server is not None:
            kill(self.task_server)
            
    def storage_path(self, *args):
        return os.path.join(self.storage_dir, *args)
    
    def copy_to_storage( self, path):
        dst = os.path.join( self.storage_dir, os.path.basename(path) )
        shutil.copy(path, dst)
        return os.path.basename(path)
    
    def get_from_storage(self, loc):
        dst = os.path.join( self.storage_dir, loc )
        return dst


class S3ServerTest(unittest.TestCase):

    def setUp(self):
        self.addCleanup(self.cleanup)

        if not os.path.exists("./test_tmp"):
            os.mkdir("test_tmp")
        self.task_server = None
        f, db_path = tempfile.mkstemp(dir="./test_tmp", prefix="tes_task_db.")
        os.close(f)
        self.storage_dir = db_path + ".storage"
        self.output_dir = db_path + ".output"
        os.mkdir(self.storage_dir)
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
        self.s3_server = popen(cmd)

        # Build server config file (YAML)
        configFile = temp_config({
            "ServerAddress": "localhost:9090",
            "DBPath": db_path,
            "WorkDir": "test_tmp",
            "Storage": [{
                "S3": {
                    "Endpoint": S3_ENDPOINT,
                    "Key": S3_ACCESS_KEY,
                    "Secret": S3_SECRET_KEY,
                }
            }]
        })

        # Start server
        cmd = ["./bin/tes-server", "-config", configFile.name]
        logging.info("Running %s" % (" ".join(cmd)))
        self.task_server = popen(cmd)
        time.sleep(5)
        
        # TES client
        self.tes = py_tes.TES("http://localhost:8000",
            "http://" + S3_ENDPOINT,
            S3_ACCESS_KEY,
            S3_SECRET_KEY)

        if not self.tes.bucket_exists(BUCKET_NAME):
            self.tes.make_bucket(BUCKET_NAME)

    # We're using this instead of tearDown because python doesn't call tearDown
    # if setUp fails. Since our setUp is complex, that means things don't get
    # properly cleaned up (e.g. processes are orphaned).
    def cleanup(self):
        if self.s3_server is not None:
            kill(self.s3_server)
            cmd = ["docker", "kill", "tes_minio_test"]
            logging.info("Running %s" % (" ".join(cmd)))
            
            cmd = ["docker", "rm", "-fv", "tes_minio_test"]
            logging.info("Running %s" % (" ".join(cmd)))
            
            popen(cmd).communicate()
        
        if self.task_server is not None:
            kill(self.task_server)
            
    def get_storage_url(self, path):
        dstpath = "s3://%s/%s" % (BUCKET_NAME, os.path.join(self.dir_name, os.path.basename(path)))
        return dstpath
    
    def copy_to_storage( self, path):
        dstpath = "s3://%s/%s" % (BUCKET_NAME, os.path.join(self.dir_name, os.path.basename(path)))
        print "uploading:", dstpath
        self.tes.upload_file(path, dstpath)
        return dstpath
    
    def get_from_storage(self, loc):
        dst = os.path.join(self.output_dir, os.path.basename(loc))
        print "Downloading %s to %s" % (loc, dst)
        self.tes.download_file(dst, loc)
        return dst
