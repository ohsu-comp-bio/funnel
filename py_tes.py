
import itertools
import re
import json
import time
import urllib2
from urlparse import urlparse
from minio import Minio


class TES:

    def __init__(self,
                 url,
                 s3_endpoint,
                 s3_access_key=None,
                 s3_secret_key=None):
        self.url = url
        self.s3_endpoint = s3_endpoint
        self.s3_access_key = s3_access_key
        self.s3_secret_key = s3_secret_key

    def get_service_info(self):
        req = urllib2.Request("%s/v1/jobs-service" % (self.url))
        u = urllib2.urlopen(req)
        return json.loads(u.read())

    def s3connect(self):
        n = urlparse(self.s3_endpoint)
        tls = None
        if n.scheme == "http":
            tls = False
        if n.scheme == "https":
            tls = True
        minioClient = Minio(
            n.netloc,
            access_key=self.s3_access_key,
            secret_key=self.s3_secret_key,
            secure=tls
        )
        return minioClient

    def upload_file(self, path, storage):
        n = urlparse(storage)
        if n.scheme != "s3":
            raise Exception("Not S3 URL")
        c = self.s3connect()
        object_name = n.path
        object_name = re.sub("^/", "", object_name)
        c.fput_object(n.netloc, object_name, path)

    def download_file(self, path, storage):
        n = urlparse(storage)
        if n.scheme != "s3":
            raise Exception("Not S3 URL")
        c = self.s3connect()
        object_name = n.path
        object_name = re.sub("^/", "", object_name)
        c.fget_object(n.netloc, object_name, path)

    def list(self, bucket):
        c = self.s3connect()
        for i in c.list_objects(bucket):
            yield "s3://%s/%s" % (bucket, i.object_name)

    def make_bucket(self, bucket):
        c = self.s3connect()
        c.make_bucket(bucket)

    def bucket_exists(self, bucket):
        c = self.s3connect()
        c.bucket_exists(bucket)

    def submit(self, task):
        req = urllib2.Request("%s/v1/jobs" % (self.url))
        u = urllib2.urlopen(req, json.dumps(task))
        data = json.loads(u.read())
        job_id = data['value']
        return job_id

    def wait(self, job_id, timeout=10):
        data = {}
        for i in itertools.count():
            if timeout > 0 and i == timeout:
                break
            req = urllib2.Request("%s/v1/jobs/%s" % (self.url, job_id))
            r = urllib2.urlopen(req)
            data = json.loads(r.read())
            if data["state"] not in ['Queued', "Running"]:
                break
            time.sleep(1)
        return data
