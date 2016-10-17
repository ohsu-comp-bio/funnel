#!/usr/bin/env python

import json
import time
import requests
import argparse

if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("-s", "--server", default="http://localhost:8000")
    parser.add_argument("task")
    args = parser.parse_args()
    
    with open(args.task) as handle:
        task = json.loads(handle.read())
    
    r = requests.post("%s/v1/jobs" % (args.server), json=task)
    data = r.json()
    print data
    job_id = data['value']

    for i in range(10):
        r = requests.get("%s/v1/jobs/%s" % (args.server, job_id))
        data = r.json()
        if data["state"] not in ['Queued', "Running"]:
            break
        time.sleep(1)
    print data
    


