FROM golang:alpine
ADD argo-messaging-linux-static /home/argo/argo-messaging
ADD config.json /home/argo/config.json
ADD host.crt /home/argo/host.crt
ADD host.key /home/argo/host.key
ENV HOME /home/argo
EXPOSE 8080
WORKDIR /home/argo
# Kafka and Zookeper take some time to boot up.
# We wait for 20'' and then start the service.
CMD sleep 20 && /home/argo/argo-messaging