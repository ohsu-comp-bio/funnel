from __future__ import print_function

import json
import os
from common_test_util import SimpleServerTest, get_abspath


class TestFileOP(SimpleServerTest):

    def test_file_mount(self):

        self.copy_to_storage(get_abspath("test_data.1"))

        task = {"name": "TestMD5",
                "project": "MyProject",
                "description": "My Desc",
                "inputs": [{"name": "infile",
                            "description": "File to be MD5ed",
                            "url": 'file://' +
                            self.storage_path('test_data.1'),
                            "type": "FILE",
                            "path": "/tmp/test_file"}],
                "outputs": [{"url": 'file://' +
                             self.storage_path('output/test_data.out'),
                             "type": "FILE",
                             "path": "/tmp/test_out"}],
                "resources": {},
                "executors": [{"image_name": "ubuntu",
                               "cmd": ["md5sum",
                                       "/tmp/test_file"],
                               "stdout": "/tmp/test_out"}]}
        print(json.dumps(task))
        task_id = self.tes.submit(task)
        self.tes.wait(task_id)

        path = self.get_from_storage('output/test_data.out')
        with open(path) as handle:
            t = handle.read()
            i = t.split()
            assert(i[0] == "fc69a359565f35bf130a127ae2ebf2da")

    def test_hard_link(self):

        self.copy_to_storage(get_abspath("test_data.1"))
        before_src_info = os.stat(self.storage_path('test_data.1'))

        task = {"name": "Hardlink test",
                "project": "MyProject",
                "description": "My Desc",
                "inputs": [{"name": "infile",
                            "description": "File to be MD5ed",
                            "url": 'file://' +
                            self.storage_path('test_data.1'),
                            "type": "FILE",
                            "path": "/tmp/test_hl_file"}],
                "resources": {},
                "executors": [{"image_name": "ubuntu",
                               "cmd": ["md5sum",
                                       "/tmp/test_hl_file"],
                               "workdir": "/workdir",
                               "stdout": "/workdir/test_out"}]}

        task_id = self.tes.submit(task)
        self.tes.wait(task_id)

        after_src_info = os.stat(self.storage_path('test_data.1'))

        assert before_src_info.st_nlink == 1
        assert after_src_info.st_nlink == 2

    def test_good_symlink(self):

        self.copy_to_storage(get_abspath("test_data.1"))
        os.symlink(self.storage_path('test_data.1'),
                   self.storage_path('test_symlink'))

        task = {"name": "TestMD5",
                "project": "MyProject",
                "description": "My Desc",
                "inputs": [{"name": "infile",
                            "description": "File to be MD5ed",
                            "url": 'file://' +
                            self.storage_path('test_symlink'),
                            "type": "FILE",
                            "path": "/tmp/test_file"}],
                "resources": {},
                "executors": [{"image_name": "ubuntu",
                               "cmd": ["md5sum",
                                       "/tmp/test_file"],
                               "workdir": "/workdir",
                               "stdout": "/workdir/test_out"}]}

        task_id = self.tes.submit(task)
        print(task_id)
        print(json.dumps(self.tes.get_task(task_id)))
        data = self.tes.wait(task_id)
        print(json.dumps(data))
        assert data['state'] == "COMPLETE"

    def test_bad_symlink(self):

        self.copy_to_storage(get_abspath("test_data.1"))
        os.symlink(self.storage_path('test_data.1'),
                   self.storage_path('test_symlink'))
        os.remove(self.storage_path('test_data.1'))

        task = {"name": "TestMD5",
                "project": "MyProject",
                "description": "My Desc",
                "inputs": [{"name": "infile",
                            "description": "File to be MD5ed",
                            "url": 'file://' +
                            self.storage_path('test_symlink'),
                            "type": "FILE",
                            "path": "/tmp/test_file"}],
                "resources": {},
                "executors": [{"image_name": "ubuntu",
                               "cmd": ["md5sum",
                                       "/tmp/test_file"],
                               "workdir": "/workdir",
                               "stdout": "/workdir/test_out"}]}

        task_id = self.tes.submit(task)
        data = self.tes.wait(task_id)
        assert data['state'] == "ERROR"

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
            "outputs": [
                {
                    "url": "file://" + self.storage_path("out-sym"),
                    "type": "FILE",
                    "path": "/tmp/sym",
                },
                {
                    "url": "file://" + self.storage_path("out-dir"),
                    "type": "DIRECTORY",
                    "path": "/tmp",
                },
            ],
            "resources": {},
            "executors": [{
                "image_name": "alpine",
                "cmd": [
                    "sh", "-c",
                    "echo foo > /tmp/foo && ln -s /tmp/foo /tmp/sym"
                ],
            }],
        }

        task_id = self.tes.submit(task)
        data = self.tes.wait(task_id)
        print(json.dumps(data))
        assert data["state"] != "ERROR"
        with open(self.storage_path("out-dir", "foo")) as fh:
            assert fh.read() == "foo\n"
        with open(self.storage_path("out-sym")) as fh:
            assert fh.read() == "foo\n"
        with open(self.storage_path("out-dir", "sym")) as fh:
            assert fh.read() == "foo\n"
