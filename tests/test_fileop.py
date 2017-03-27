#!/usr/bin/env python

import os

from common_test_util import SimpleServerTest, get_abspath


class TestFileOP(SimpleServerTest):

    def test_file_mount(self):

        self.copy_to_storage(get_abspath("test_data.1"))

        task = {"name": "TestMD5",
                "projectId": "MyProject",
                "description": "My Desc",
                "inputs": [{"name": "infile",
                            "description": "File to be MD5ed",
                            "location": 'file://' +
                            self.storage_path('test_data.1'),
                            "class": "File",
                            "path": "/tmp/test_file"}],
                "outputs": [{"location": 'file://' +
                             self.storage_path('output/test_data.out'),
                             "class": "File",
                             "path": "/tmp/test_out"}],
                "resources": {"volumes": [{"name": "test_disk",
                                           "sizeGb": 5,
                                           "mountPoint": "/tmp"}]},
                "docker": [{"imageName": "ubuntu",
                            "cmd": ["md5sum",
                                    "/tmp/test_file"],
                            "stdout": "/tmp/test_out"}]}

        job_id = self.tes.submit(task)
        self.tes.wait(job_id)

        path = self.get_from_storage('output/test_data.out')
        with open(path) as handle:
            t = handle.read()
            i = t.split()
            assert(i[0] == "fc69a359565f35bf130a127ae2ebf2da")

    def test_hard_link(self):

        self.copy_to_storage(get_abspath("test_data.1"))
        before_src_info = os.stat(self.storage_path('test_data.1'))

        task = {"name": "TestMD5",
                "projectId": "MyProject",
                "description": "My Desc",
                "inputs": [{"name": "infile",
                            "description": "File to be MD5ed",
                            "location": 'file://' +
                            self.storage_path('test_data.1'),
                            "class": "File",
                            "path": "/tmp/test_file"}],
                "resources": {
                    "volumes": [{"name": "test_disk",
                                 "sizeGb": 5,
                                 "mountPoint": "/tmp",
                                 "readonly": True}]
                },
                "docker": [{"imageName": "ubuntu",
                            "cmd": ["md5sum",
                                    "/tmp/test_file"],
                            "workdir": "/workdir",
                            "stdout": "/workdir/test_out"}]}

        job_id = self.tes.submit(task)
        self.tes.wait(job_id)

        after_src_info = os.stat(self.storage_path('test_data.1'))

        assert before_src_info.st_nlink == 1
        assert after_src_info.st_nlink == 2

    def test_good_symlink(self):

        self.copy_to_storage(get_abspath("test_data.1"))
        os.symlink(self.storage_path('test_data.1'),
                   self.storage_path('test_symlink'))

        task = {"name": "TestMD5",
                "projectId": "MyProject",
                "description": "My Desc",
                "inputs": [{"name": "infile",
                            "description": "File to be MD5ed",
                            "location": 'file://' +
                            self.storage_path('test_symlink'),
                            "class": "File",
                            "path": "/tmp/test_file"}],
                "resources": {
                    "volumes": [{"name": "test_disk",
                                 "sizeGb": 5,
                                 "mountPoint": "/tmp",
                                 "readonly": True}]
                },
                "docker": [{"imageName": "ubuntu",
                            "cmd": ["md5sum",
                                    "/tmp/test_file"],
                            "workdir": "/workdir",
                            "stdout": "/workdir/test_out"}]}

        job_id = self.tes.submit(task)
        data = self.tes.wait(job_id)
        assert data['state'] == "Complete"

    def test_bad_symlink(self):

        self.copy_to_storage(get_abspath("test_data.1"))
        os.symlink(self.storage_path('test_data.1'),
                   self.storage_path('test_symlink'))
        os.remove(self.storage_path('test_data.1'))

        task = {"name": "TestMD5",
                "projectId": "MyProject",
                "description": "My Desc",
                "inputs": [{"name": "infile",
                            "description": "File to be MD5ed",
                            "location": 'file://' +
                            self.storage_path('test_symlink'),
                            "class": "File",
                            "path": "/tmp/test_file"}],
                "resources": {
                    "volumes": [{"name": "test_disk",
                                 "sizeGb": 5,
                                 "mountPoint": "/tmp",
                                 "readonly": True}]
                },
                "docker": [{"imageName": "ubuntu",
                            "cmd": ["md5sum",
                                    "/tmp/test_file"],
                            "workdir": "/workdir",
                            "stdout": "/workdir/test_out"}]}

        job_id = self.tes.submit(task)
        data = self.tes.wait(job_id)
        assert data['state'] == "Error"

    def test_no_output_in_readonly(self):

        self.copy_to_storage(get_abspath("test_data.1"))
        os.symlink(self.storage_path('test_data.1'),
                   self.storage_path('test_symlink'))
        os.remove(self.storage_path('test_data.1'))

        task = {"name": "TestMD5",
                "projectId": "MyProject",
                "description": "My Desc",
                "inputs": [{"name": "infile",
                            "description": "File to be MD5ed",
                            "location": 'file://' +
                            self.storage_path('test_symlink'),
                            "class": "File",
                            "path": "/tmp/test_file"}],
                "outputs": [{"location": 'file://' +
                             self.storage_path('output/test_out'),
                             "class": "File",
                             "path": "/tmp/test_out"}],
                "resources": {
                    "volumes": [{"name": "test_disk",
                                 "sizeGb": 5,
                                 "mountPoint": "/tmp",
                                 "readonly": True}]
                },
                "docker": [{"imageName": "ubuntu",
                            "cmd": ["md5sum",
                                    "/tmp/test_file"],
                            "workdir": "/tmp",
                            "stdout": "/tmp/test_out"}]}

        job_id = self.tes.submit(task)
        data = self.tes.wait(job_id)
        assert data['state'] == "Error"

    def test_symlink_in_output(self):
        """
        test_symlink_in_output

        Test the case where a container creates a symlink in an output path.
        From the view of the host system where Funnel is running, this creates
        a broken link, because the source of the symlink is a path relative
        to the container filesystem.

        Funnel can fix some of these cases using volume definitions, which
        is being tested here.
        """
        task = {
            "name": "Test symlink in output",
            "outputs": [{
                "location": "file://" + self.storage_path("out"),
                "class": "File",
                "path": "/tmp/out",
            }],
            "resources": {
                "volumes": [{
                    "name": "testvol",
                    "sizeGb": 5,
                    "mountPoint": "/tmp",
                }],
            },
            "docker": [{
                "imageName": "alpine",
                "cmd": [
                    "sh", "-c",
                    "echo foo > /tmp/foo && ln -s /tmp/foo /tmp/out"
                ],
            }],
        }
        job_id = self.tes.submit(task)
        data = self.tes.wait(job_id)
        print data
        assert data["state"] != "Error"
        with open(self.storage_path("out")) as fh:
            assert fh.read() == "foo\n"
