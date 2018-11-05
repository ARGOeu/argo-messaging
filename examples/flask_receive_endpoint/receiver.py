#!/usr/bin/env python

# Example of a remote endpoint used to receive push messages
# The endpoint is a simple flask app that by default listens to port 5000
# It receives push messages that are delivered with http POST to `host.remote.node:5000/receive_here`
# It dumps the message properties and the decoded payload to a local file `./flask_receiver.log`
#
# To run the example endpoint issue:
#  $ export FLASK_APP=receiver.py
#  $ flask run
#
# If you want the endpoint to support https issue:
#  $ ./receiver.py --cert /path/to/cert --key /path/to/cert/key
#
# You can also specify the bind port with the -port argument, default is 5000

from flask import Flask
from flask import request
import argparse
import json
import base64
from logging.config import dictConfig
import ssl
import flask_cors
from flask.logging import default_handler

dictConfig({
    'version': 1,
    'formatters': {'default': {
        'format': '[%(asctime)s] %(levelname)s in %(module)s: %(message)s',
    }},
    'handlers': {
        'wsgi': {
            'class': 'logging.StreamHandler',
            'stream': 'ext://flask.logging.wsgi_errors_stream',
            'formatter': 'default',
            'level': 'INFO'
        },
        'logfile': {
            'class': 'logging.FileHandler',
            'filename': 'flask_receiver.log',
            'formatter': 'default',
            'level': 'INFO'
        }
    },
    'root': {
        'level': 'INFO',
        'handlers': ['wsgi', 'logfile']
    }
})


app = Flask(__name__)

app.logger.removeHandler(default_handler)


@app.route('/receive_here', methods=['POST'])
def receive_msg():

    try:
        data = json.loads(request.get_data())
        msg = data["message"]
        msg_json = json.dumps(data, indent=4)

        if "subscription" not in data:
            raise KeyError("subscription field missing from request data: {}".format(msg_json))

        if "messageId" not in msg:
            raise KeyError("messageId field missing from request message: {}".format(msg_json))

        if "data" not in msg:
            raise KeyError("messageId field missing from request message: {}".format(msg_json))

        app.logger.info(data)

        return 'Message received', 201

    except Exception as e:
        app.logger.error(e.message)
        return e.message, 400


if __name__ == "__main__":

    parser = argparse.ArgumentParser(description="Simple flask endpoint for push subscriptions")

    parser.add_argument(
        "-cert", "--cert", metavar="STRING", help="Certificate location",
        default="/etc/grid-security/hostcert.pem", dest="cert")

    parser.add_argument(
        "-key", "--key", metavar="STRING", help="Key location",
        default="/etc/grid-security/hostkey.pem", dest="key")

    parser.add_argument(
        "-port", "--port", metavar="INTEGER", help="Bind port",
        default=5000, type=int, dest="port")

    args = parser.parse_args()

    flask_cors.CORS(app=app, methods=["OPTIONS", "HEAD", "POST"], allow_headers=["X-Requested-With", "Content-Type"])

    context = ssl.SSLContext(ssl.PROTOCOL_TLSv1_2)
    context.load_cert_chain(args.cert, args.key)

    app.run(host='0.0.0.0', port=args.port, ssl_context=context, threaded=True, debug=True)
