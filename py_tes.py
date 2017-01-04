
import itertools
import re
import os
import json
import time
import urllib2
import argparse
from urlparse import urlparse
import jwt
from minio import Minio
from minio.error import ResponseError

class TES:
    def __init__(self, url, s3_access_key=None, s3_secret_key=None, secret=None):
        self.url = url
        info = self.get_service_info()
        if 's3.endpoint' in info['storageConfig']:
            self.s3endpoint = info['storageConfig']['s3.endpoint']
            if s3_access_key is not None:
                self.s3_access_key = s3_access_key
            else:
                self.s3_access_key = os.environ.get("AWS_ACCESS_KEY_ID", None)
            if s3_secret_key is not None:
                self.s3_secret_key = s3_secret_key
            else:
                self.s3_secret_key = os.environ.get("AWS_SECRET_ACCESS_KEY", None)
        else:
            self.s3endpoint = None
        self.secret = secret
    
    def get_service_info(self):
        req = urllib2.Request("%s/v1/jobs-service" % (self.url))
        u = urllib2.urlopen(req)
        return json.loads(u.read())
    
    def s3connect(self):
        n = urlparse(self.s3endpoint)
        tls = None
        if n.scheme == "http":
            tls = False
        if n.scheme == "https":
            tls = True
        minioClient = Minio(n.netloc,
                    access_key=self.s3_access_key,
                    secret_key=self.s3_secret_key,
                    secure=tls)
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
        encoded = None
        if self.secret is not None and self.s3_access_key is not None and self.s3_secret_key is not None:
            encoded = jwt.encode({'AWS_ACCESS_KEY_ID': self.s3_access_key, 'AWS_SECRET_ACCESS_KEY' : self.s3_secret_key}, self.secret, algorithm='HS256')
        req = urllib2.Request("%s/v1/jobs" % (self.url))
        if encoded is not None:
            req.add_header('authorization', "JWT %s" % (encoded))
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

