#!/usr/bin/env python

from common_test_util import S3ServerTest, get_abspath


class TestFileOP(S3ServerTest):

    def test_file_mount(self):

        in_loc = self.copy_to_storage(get_abspath("test_data.1"))
        out_loc = self.get_storage_url("test_data.out")

        task = {
            "name": "TestMD5",
            "projectId": "MyProject",
            "description": "My Desc",
            "inputs": [
                {
                    "name": "infile",
                    "description": "File to be MD5ed",
                    "location": in_loc,
                    "class": "File",
                    "path": "/tmp/test_file"
                }
            ],
            "outputs": [
                {
                    "location": out_loc,
                    "class": "File",
                    "path": "/tmp/test_out"
                }
            ],
            "resources": {
                "volumes": [{
                    "name": "test_disk",
                    "sizeGb": 5,
                    "mountPoint": "/tmp"
                }]
            },
            "docker": [
                {
                    "imageName": "ubuntu",
                    "cmd": ["md5sum", "/tmp/test_file"],
                    "stdout": "/tmp/test_out"
                }
            ]
        }

        job_id = self.tes.submit(task)
        self.tes.wait(job_id)

        path = self.get_from_storage(out_loc)
        with open(path) as handle:
            t = handle.read()
            i = t.split()
            assert(i[0] == "fc69a359565f35bf130a127ae2ebf2da")
