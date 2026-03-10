#!/bin/sh

(nohup /usr/local/bin/dockerd-entrypoint.sh --experimental  &) 2> /dev/null

timeout=30
while [ ! -f /var/run/docker.pid ]; do
    if [ "$timeout" == 0 ]; then
        echo "ERROR: docker failed to start within timeout"
        exit 1
    fi
    sleep 1
    timeout=$(($timeout - 1))
done

$@
