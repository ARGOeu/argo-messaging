#!/usr/bin/env python

# Example of a remote endpoint used to receive push messages
# The endpoint is a simple flask app that by default listens to port 5000
# It receives push messages that are delivered with http POST to `host.remote.node:5000/receive_here`
# It dumps the message properties and the decoded payload to a local file `./api.dmp`
#
# To run the example endpoint issue:
#  $ export FLASK_APP=receiver.py
#  $ flask run


from flask import Flask
from flask import request
import json
import base64
app = Flask(__name__)


@app.route('/receive_here', methods=['POST'])
def receive_msg():
    try:
        data = json.loads(request.get_data())
        msg = data["message"]
        with open("api.dump", "a") as fo:
            fo.write("subscription:" + str(data["subscription"] + "\n"))
            fo.write("message_id:" + str(msg["message_id"]) + "\n")
            fo.write("attributes:" + str(msg["attributes"]) + "\n")
            fo.write("data:" + str(base64.b64decode(msg["data"]))+ "\n")
            fo.write("------\n")
        return '', 201

    except (KeyError, ValueError):
        return '', 400
