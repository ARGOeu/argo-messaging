#!/usr/bin/env python

# Example of a remote endpoint used to receive push messages
# The endpoint is a simple flask app that by default listens to port 5000
# It receives push messages that are delivered with http POST to `host:port/receive_here`
# It dumps the message properties and the decoded payload to a local file `./flask_receiver.log`
#
# It allows for user to view the received messages since the endpoint started by visiting `host:port/messages`
# page in the browser. The page auto-refreshes automatically every 3 seconds
#
# To run the example endpoint issue:
#  $ export FLASK_APP=receiver.py
#  $ flask run
#
# If you want the endpoint to support https issue:
#  $ ./receiver.py --cert /path/to/cert --key /path/to/cert/key
#
# You can specify the hostname to listen on with the -host argument, default is 127.0.0.0 
# You can also specify the bind port with the -port argument, default is 5000.
# Lastly, you can also specify which message format the endpoint should expect
# --single or --multiple.
#
# The --authorization-header parameter allows to define an expected authorization has supplied 
# from the ams push server in an authorization header. This is used to validate that you are indeed
# receiving requests from the designated ams push server
#
# The --verification-hash parameter allows to supply a verification hash to the endpoint in order 
# to be verified by the AMS (see ams documentation for verification of remote endpoint in push configruations)

from re import A
from flask import Flask
from flask import request
from flask import Response
from datetime import datetime
import base64
import argparse
import json
from logging.config import dictConfig
import ssl
import flask_cors
from flask.logging import default_handler
from string import Template

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

RECEIVED_MSGS=[]

AUTHZ_HEADER = ""
START_TIME = datetime.utcnow().isoformat()

app = Flask(__name__)


app.logger.removeHandler(default_handler)

# decode a received msg payload and create a 3-tuple with msgid, publishtime and text payload
def msg_to_tuple(msg):
    msg_decoded = base64.b64decode(msg.get("data")).decode('utf-8')
    pub_time = msg.get("publishTime")
    msg_ID = msg.get("messageId")
    return (msg_ID, pub_time, msg_decoded)
    

@app.route('/receive_here', methods=['POST'])
def receive_msg():

    # if there is an authz header provided, check it
    if AUTHZ_HEADER != "":
        print(request.headers.get("Authorization"))
        if request.headers.get("Authorization") != AUTHZ_HEADER:
            return "UNAUTHORIZED", 401
        
    if MESSAGE_FORMAT == "single":

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

            RECEIVED_MSGS.append(msg_to_tuple(msg))
            
            app.logger.info(data)
            

            return 'Message received', 201

        except Exception as e:
            app.logger.error(e.message)
            return e.message, 400

    elif MESSAGE_FORMAT == "multi":

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
                    
                RECEIVED_MSGS.append(msg_to_tuple(msg))
            
            app.logger.info(data)

            return 'Messages received', 201

        except Exception as e:
            app.logger.error(e.message)
            return e.message, 400


@app.route('/messages', methods=['GET'])
def view_msg():
    p = Template('<p>Received messages <em>(since $tm)</em>:</p>').substitute(tm=START_TIME)
    meta = '<meta http-equiv="refresh" content="3" />'
    ul = ''
    for msg in RECEIVED_MSGS:
       li = Template('<li><p>$id - <sub>$pub</sub></br>$msg</p></li>').substitute(id=msg[0],pub=msg[1],msg=msg[2])
       ul = ul + li
    html = Template('<html><head><title>Received Messages</title> $meta</head><body>$par<ul>$ul</ul></body></html>').substitute(par=p,meta=meta,ul=ul)
    return Response(response=html, status=200, content_type="text/html")


@app.route('/ams_verification_hash', methods=['GET'])
def return_verification_hash():
    return Response(response=VERIFICATION_HASH, status=200, content_type="plain/text")


if __name__ == "__main__":

    parser = argparse.ArgumentParser(description="Simple flask endpoint for push subscriptions")

    parser.add_argument(
        "-host", "--host", metavar="STRING", help="Hostname to listen to",
        default="127.0.0.1", dest="host")

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

    parser.add_argument(
        "-ah", "--authorization-header", metavar="STRING", help="Expected authorization header",
        required=False, dest="authz")

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

    AUTHZ_HEADER = args.authz

    if args.single_message:
        MESSAGE_FORMAT = "single"

    if args.multi_message:
        MESSAGE_FORMAT = "multi"

    app.run(host=args.host, port=args.port, ssl_context=context, threaded=True, debug=True)

