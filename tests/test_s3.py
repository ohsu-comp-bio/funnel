#!/usr/bin/env python

from common_test_util import S3ServerTest, get_abspath


class TestS3(S3ServerTest):

    def test_file_mount(self):

        in_loc = self.copy_to_storage(get_abspath("test_data.1"))
        out_loc = self.get_storage_url("test_data.out")

        task = {
            "name": "TestMD5",
            "project": "MyProject",
            "description": "My Desc",
            "inputs": [
                {
                    "name": "infile",
                    "description": "File to be MD5ed",
                    "url": in_loc,
                    "type": "FILE",
                    "path": "/tmp/test_file"
                }
            ],
            "outputs": [
                {
                    "url": out_loc,
                    "type": "FILE",
                    "path": "/tmp/test_out"
                }
            ],
            "resources": {},
            "executors": [
                {
                    "image_name": "ubuntu",
                    "cmd": ["md5sum", "/tmp/test_file"],
                    "stdout": "/tmp/test_out"
                }
            ]
        }

        task_id = self.tes.submit(task)
        self.tes.wait(task_id, timeout=20)

        path = self.get_from_storage(out_loc)
        with open(path) as handle:
            t = handle.read()
            i = t.split()
            assert(i[0] == "fc69a359565f35bf130a127ae2ebf2da")
