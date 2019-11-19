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
# Lastly, you can also specify which message format the endpoint should expect
# --single or --multiple

from flask import Flask
from flask import request
from flask import Response
import argparse
import json
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

VERIFICATION_HASH = ""

MESSAGE_FORMAT = ""

app = Flask(__name__)

app.logger.removeHandler(default_handler)


@app.route('/receive_here', methods=['POST'])
def receive_msg():

    if MESSAGE_FORMAT is "single":

        try:
            data = json.loads(request.get_data())

            data_json = json.dumps(data, indent=4)

            if "message" not in data:
                raise KeyError("message field missing from request data: {}".format(data_json))

            if "subscription" not in data:
                raise KeyError("subscription field missing from request data: {}".format(msg_json))

            msg = data["message"]

            msg_json = json.dumps(data, indent=4)

            if "messageId" not in msg:
                raise KeyError("messageId field missing from request message: {}".format(msg_json))

            if "data" not in msg:
                raise KeyError("data field missing from request message: {}".format(msg_json))

            app.logger.info(data)

            return 'Message received', 201

        except Exception as e:
            app.logger.error(e.message)
            return e.message, 400

    elif MESSAGE_FORMAT is "multi":

        try:
            data = json.loads(request.get_data())

            data_json = json.dumps(data, indent=4)

            if "messages" not in data:
                raise KeyError("messages field missing from request data: {}".format(data_json))

            messages = data["messages"]

            for datum in messages:

                msg_json = json.dumps(datum, indent=4)

                if "message" not in datum:
                    raise KeyError("message field missing from request data: {}".format(msg_json))

                if "subscription" not in datum:
                    raise KeyError("subscription field missing from request data: {}".format(msg_json))

                msg = datum["message"]

                if "messageId" not in msg:
                    raise KeyError("messageId field missing from request message: {}".format(msg_json))

                if "data" not in msg:
                    raise KeyError("data field missing from request message: {}".format(msg_json))

            app.logger.info(data)

            return 'Messages received', 201

        except Exception as e:
            app.logger.error(e.message)
            return e.message, 400


@app.route('/ams_verification_hash', methods=['GET'])
def return_verification_hash():
    return Response(response=VERIFICATION_HASH, status=200, content_type="plain/text")


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

    parser.add_argument(
        "-vh", "--verification-hash", metavar="STRING", help="Verification hash for the push endpoint",
        required=True, dest="vhash")

    group = parser.add_mutually_exclusive_group(required=True)

    group.add_argument("--single", action="store_true", help="The endpoint should expect single message format",
                       dest="single_message")

    group.add_argument("--multiple", action="store_true", help="The endpoint should expect multiple messages format",
                       dest="multi_message")

    args = parser.parse_args()

    flask_cors.CORS(app=app, methods=["OPTIONS", "HEAD", "POST"], allow_headers=["X-Requested-With", "Content-Type"])

    context = ssl.SSLContext(ssl.PROTOCOL_TLSv1_2)
    context.load_cert_chain(args.cert, args.key)

    VERIFICATION_HASH = args.vhash

    if args.single_message:
        MESSAGE_FORMAT = "single"

    if args.multi_message:
        MESSAGE_FORMAT = "multi"

    app.run(host='0.0.0.0', port=args.port, ssl_context=context, threaded=True, debug=True)

