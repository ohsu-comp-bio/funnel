from __future__ import print_function

import logging
import os
import py_tes
import shutil
import signal
import subprocess
import tempfile
import time
import unittest
import requests
import polling
import yaml
import docker


S3_ENDPOINT = "localhost:9000"
S3_ACCESS_KEY = "AKIAIOSFODNN7EXAMPLE"
S3_SECRET_KEY = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
BUCKET_NAME = "tes-test"
WORK_DIR = os.path.join(os.path.dirname(os.path.dirname(__file__)),
                        "test_work")


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


def config_seconds(sec):
    # The funnel config is currently parsed as nanoseconds
    # this helper makes that manageale
    return int(sec * 1000000000)


class SimpleServerTest(unittest.TestCase):

    def setUp(self):
        self.addCleanup(self.cleanup)
        if not os.path.exists("./test_tmp"):
            os.mkdir("test_tmp")
        self.task_server = None
        f, db_path = tempfile.mkstemp(dir="./test_tmp", prefix="tes_task_db.")
        os.close(f)
        self.storage_dir = os.path.abspath(db_path + ".storage")
        self.funnel_work_dir = os.path.abspath(db_path + ".work-dir")
        os.mkdir(self.storage_dir)
        os.mkdir(self.funnel_work_dir)

        # Build server config file (YAML)
        rate = config_seconds(0.05)
        configFile = temp_config({
            "HostName": "localhost",
            "HTTPPort": "8000",
            "RPCPort": "9090",
            "DBPath": db_path,
            "WorkDir": self.funnel_work_dir,
            "Storage": [{
                "Local": {
                    "AllowedDirs": [self.storage_dir]
                }
            }],
            "LogLevel": "debug",
            "Worker": {
                "Timeout": -1,
                "StatusPollRate": rate,
                "LogUpdateRate": rate,
                "NewJobPollRate": rate,
                "UpdateRate": rate,
                "TrackerRate": rate
            },
            "ScheduleRate": rate,
        })

        # Start server
        cmd = ["funnel", "server", "--config", configFile.name]
        logging.info("Running %s" % (" ".join(cmd)))
        self.task_server = popen(cmd)
        signal.signal(signal.SIGINT, self.cleanup)
        time.sleep(1)
        self.tes = py_tes.TES("http://localhost:8000")

    # We're using this instead of tearDown because python doesn't call tearDown
    # if setUp fails. Since our setUp is complex, that means things don't get
    # properly cleaned up (e.g. processes are orphaned).
    def cleanup(self, *args):
        if self.task_server is not None:
            kill(self.task_server)

    def storage_path(self, *args):
        return os.path.join(self.storage_dir, *args)

    def copy_to_storage(self, path):
        dst = os.path.join(self.storage_dir, os.path.basename(path))
        shutil.copy(path, dst)
        return os.path.basename(path)

    def get_from_storage(self, loc):
        dst = os.path.join(self.storage_dir, loc)
        return dst

    def wait_for_container(self, name, timeout=5):
        dclient = docker.from_env()

        def on_poll():
            try:
                dclient.containers.get(name)
                return True
            except BaseException:
                return False
        polling.poll(on_poll, timeout=timeout, step=0.1)

    def wait_for_container_stop(self, name, timeout=5):
        dclient = docker.from_env()

        def on_poll():
            try:
                dclient.containers.get(name)
                return False
            except BaseException:
                return True
        polling.poll(on_poll, timeout=timeout, step=0.1)

    def wait(self, key, timeout=5):
        """
        Waits for tes-wait to return <key>
        """
        def on_poll():
            try:
                r = requests.get("http://127.0.0.1:5000/")
                return r.status_code == 200 and r.text == key
            except requests.ConnectionError:
                return False

        polling.poll(on_poll, timeout=timeout, step=0.1)

    def resume(self):
        """
        Continue from tes-wait
        """
        requests.get("http://127.0.0.1:5000/shutdown")


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
        self.funnel_work_dir = os.path.abspath(db_path + ".work-dir")
        os.mkdir(self.storage_dir)
        os.mkdir(self.output_dir)
        os.mkdir(self.funnel_work_dir)

        self.dir_name = os.path.basename(db_path)

        # start s3 server
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
            "HostName": "localhost",
            "HTTPPort": "8000",
            "RPCPort": "9090",
            "DBPath": db_path,
            "WorkDir": self.funnel_work_dir,
            "Storage": [{
                "S3": {
                    "Endpoint": S3_ENDPOINT,
                    "Key": S3_ACCESS_KEY,
                    "Secret": S3_SECRET_KEY,
                }
            }]
        })

        # Start server
        cmd = ["funnel", "server", "--config", configFile.name]
        logging.info("Running %s" % (" ".join(cmd)))
        self.task_server = popen(cmd)
        time.sleep(5)

        # TES client
        self.tes = py_tes.TES(
            "http://localhost:8000",
            "http://" + S3_ENDPOINT,
            S3_ACCESS_KEY,
            S3_SECRET_KEY
        )

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

            popen(cmd).communicate()

            cmd = ["docker", "rm", "-fv", "tes_minio_test"]
            logging.info("Running %s" % (" ".join(cmd)))

            popen(cmd).communicate()

        if self.task_server is not None:
            kill(self.task_server)

    def get_storage_url(self, path):
        dstpath = "s3://%s/%s" % (
            BUCKET_NAME, os.path.join(self.dir_name, os.path.basename(path))
        )
        return dstpath

    def copy_to_storage(self, path):
        dstpath = "s3://%s/%s" % (
            BUCKET_NAME, os.path.join(self.dir_name, os.path.basename(path))
        )
        print("uploading:", dstpath)
        self.tes.upload_file(path, dstpath)
        return dstpath

    def get_from_storage(self, loc):
        dst = os.path.join(self.output_dir, os.path.basename(loc))
        print("Downloading %s to %s" % (loc, dst))
        self.tes.download_file(dst, loc)
        return dst
