---
id: ams_push_worker
title: AMS Push Worker 
---

AMS Push worker (ver 0.1.0) is a command line utility that let’s you simulate AMS push functionality by pulling messages from an actual AMS project/subscription and pushing them to an endpoint in your local development environment. It’s written in go and it’s packaged as a single binary with no depedencies.


- Github repo: https://github.com/ARGOeu/ams-push-worker
- Readme: https://github.com/ARGOeu/ams-push-worker#readme
- Linux Binary: https://github.com/ARGOeu/ams-push-worker/releases/download/0.1.0/ams-push-worker_linux_x86_64.zip

If you are developing an endpoint that will receive messages from AMS service you can take a look at a simple working python example at the following link:

- https://github.com/ARGOeu/argo-messaging/blob/devel/examples/flask_receive_endpoint/receiver.py

**Some more usefull links:**
- AMS docs: https://argoeu.github.io/argo-messaging
- Push enabled subscriptions: https://argoeu.github.io/argo-messaging/docs/api_subscriptions#push-enabled-subscriptions
