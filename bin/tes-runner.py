#!/usr/bin/env python

import json
import time
import urllib
import argparse

if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("-s", "--server", default="http://localhost:8000")
    parser.add_argument("task")
    args = parser.parse_args()
    
    with open(args.task) as handle:
        task = json.loads(handle.read())
    
    u = urllib.urlopen("%s/v1/jobs" % (args.server), json.dumps(task))
    data = json.loads(u.read())
    
    print data
    job_id = data['value']

    while True:
        r = urllib.urlopen("%s/v1/jobs/%s" % (args.server, job_id))
        data = json.loads(r.read())
        if data["state"] not in ['Queued', "Running"]:
            break
        time.sleep(1)
    print json.dumps(data, indent=4)
    


