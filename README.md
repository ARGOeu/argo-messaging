[![Build Status](https://travis-ci.org/ARGOeu/argo-messaging.svg?branch=devel)](https://travis-ci.org/ARGOeu/argo-messaging)
# ARGO Messaging

> ## :warning: Warning :warning:
> These installation instructions are meant for running the service for demo purposes. If you want to operate the service for anything else other than a simple demo, please implement a deployment model that meets your requirements.

In order to build, test and run the service, recent versions of the docker-engine (>=1.12) and the docker-compose (>= 1.8.0) are required. Step 1 refers to the docker installation on Ubuntu 16.04.1, please adopt accordingly your Linux distribution or OS.

## Install docker from dockerproject.org (Ubuntu 16.04.1)

```shell
$ sudo apt-key adv --keyserver hkp://pool.sks-keyservers.net:80 --recv-keys 58118E89F3A912897C070ADBF76221572C52609D
$ echo "deb https://apt.dockerproject.org/repo ubuntu-xenial main" | sudo tee /etc/apt/sources.list.d/docker.list
$ sudo apt-get update
$ sudo apt-cache policy docker-engine
$ sudo apt-get install linux-image-extra-$(uname -r) linux-image-extra-virtual
$ sudo apt-get install docker-engine
```

We advise you to follow the steps described in docker manual. For Ubuntu:

- Prerequisites : https://docs.docker.com/engine/installation/linux/ubuntulinux/#prerequisites
- Install : https://docs.docker.com/engine/installation/linux/ubuntulinux/#install
- Add a docker group [https://docs.docker.com/engine/installation/linux/ubuntulinux/#/create-a-docker-group] .

**Note:** Don't forget to login logout before running the docker as a non root user. This ensures your user is running with the correct permissions.

## Install docker-compose

We are using version of the Compose file format. To install the latest docker-compose, follow the guidelines here: https://github.com/docker/compose/releases

## Clone the argo-messaging repository

```shell
$ git clone https://github.com/ARGOeu/argo-messaging
```

## Get certificates (skip this step if you already have certificates)

The ARGO Messaging services requires certificates in order to operates. The easiest way is to get certificates from letsencrypt. You can follow the instructions from the letsencrypt website or use the docker letsencrypt docker image. One caveat of this approach is that the certificate files end up in the ```etc/live``` directory (see below) and will be owned by the root user.

```shell
$ mkdir -p ${HOME}/letsencrypt/{etc,var}
$ docker run -it --rm -p 443:443 -p 80:80 --name certbot \
    -v "$HOME/letsencrypt/etc:/etc/letsencrypt" \
    -v "$HOME/letsencrypt/var:/var/lib/letsencrypt" \
    quay.io/letsencrypt/letsencrypt:latest certonly
$ cd argo-messaging
# Comment: Please change owneship of ${HOME}/letsencrypt to your user
$ cp ${HOME}/letsencrypt/etc/live/*/fullchain.pem host.crt
$ sudo cp ${HOME}/letsencrypt/etc/live/*/privkey.pem host.key
```
## Edit the default configuration file (config.json)

In the ```argo-messaging``` directory, edit ```config.json```:

```diff
{
"bind_ip":"",
"port":8080,
-  "zookeeper_hosts":["localhost"],
-  "store_host":"localhost",
+  "zookeeper_hosts":["zookeeper"],
+  "store_host":"mongo",
"store_db":"argo_msg",
-  "certificate":"/etc/pki/tls/certs/localhost.crt",
-  "certificate_key":"/etc/pki/tls/private/localhost.key",
+  "certificate":"./host.crt",
+  "certificate_key":"./host.key",
"service_token":"CHANGE-THIS-TO-A-LONG-STRING",
"push_enabled": false
}
```

**Note:** Make sure that you change the service_token to a long string.

## Edit docker-compose.yml

In the ```argo-messaging``` directory, edit ```docker-compose.yml``` and add the public IP address of your host to the ```KAFKA_ADVERTISED_HOST_NAME``` key.

## Run the tests

```shell
$ docker run --env hostUID=`id -u`:`id -g` --rm -v "$PWD":/usr/src/myapp -w /usr/src/myapp golang:1.7 make go-test
```

## Build the service

```shell
$ docker run --env hostUID=`id -u`:`id -g` --rm -v "$PWD":/usr/src/myapp -w /usr/src/myapp golang:1.7 make go-build-linux-static
```

## Start the service

```shell
$ docker-compose build
$ docker-compose up -d
```

##  Test that the service is running

```shell
$ curl https://<HOSTNAME>/v1/projects?key=<YOUR_SERVICE_TOKEN>
```

**Note:** Change ```<HOSTNAME>``` to the hostname of your host and ```<SERVICE_TOKEN>``` to the service token that you have added in ```config.json```. You should get an empty json response:

```shell
{}
```

## Stop the service

```shell
$ docker-compose stop
```

## Congratulations!

Please visit http://argoeu.github.io/messaging/v1/ to learn how to use the service.

## Credits

The ARGO Messaging Service is developed by [GRNET](http://www.grnet.gr)

The work represented by this software was partially funded by 
 - EGI-ENGAGE project through the European Union (EU) Horizon 2020 program under Grant number 654142.
 - EOSC-Hub project through the European Union (EU) Horizon 2020 program under Grant number 77753642.
