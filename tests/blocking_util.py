#!/usr/bin/env python

"""
Simple tool to allow for blocking during TES tasks
"""

from flask import Flask, request


blocking_app = Flask(__name__)


def shutdown_server():
    func = request.environ.get('werkzeug.server.shutdown')
    if func is None:
        raise RuntimeError('Not running with the Werkzeug Server')
    func()


@blocking_app.route('/shutdown', methods=['POST', 'GET'])
def shutdown():
    shutdown_server()
    return ""


if __name__ == "__main__":
    blocking_app.run(host='0.0.0.0', port=5000)
