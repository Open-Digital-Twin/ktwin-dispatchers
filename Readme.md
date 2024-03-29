# KTwin Dispatcher Repository

This repository contains the MQTT and Cloud Event Dispatchers used to route MQTT messages and AMQP messages within RabbitMQ cluster running under Knative Broker abstraction.

KTwin uses RabbitMQ 3.12 which provides Native MQTT functionalities which allows serving millions of MQTT devices. RabbitMQ offers protocol interoperability between AMQP and MQTT. More details about how RabbitMQ Native MQTT [here](https://blog.rabbitmq.com/posts/2023/03/native-mqtt/).

The Cloud Event Dispatcher is responsible to forward Cloud Event messages from Virtual components generated within KTwin cluster to real devices connected to the cluster using MQTT. The MQTT Dispatcher forwards MQTT messages from real components to virtual services.

## Build

Build MQTT Dispatcher:

```sh
docker build -t ghcr.io/open-digital-twin/ktwin-mqtt-dispatcher:0.1 --build-arg SERVICE_NAME=mqtt-dispatcher .
```

Build Cloud Event Dispatcher:

```sh
docker build -t ghcr.io/open-digital-twin/ktwin-cloud-event-dispatcher:0.1 --build-arg SERVICE_NAME=cloud-event-dispatcher .
```

## Container Push

```sh
docker push ghcr.io/open-digital-twin/ktwin-mqtt-dispatcher:0.1
```

```sh
docker push ghcr.io/open-digital-twin/ktwin-cloud-event-dispatcher:0.1
```

## Run

Run MQTT Dispatcher:

```sh
docker run -it --rm \
    -e SERVICE_NAME=mqtt-dispatcher-1 \
    -e PROTOCOL=amqp \
    -e SERVER_URL=localhost \
    -e SERVER_PORT="5672" \
    -e USERNAME=rabbitmq \
    -e PASSWORD=rabbitmq \
    -e DECLARE_EXCHANGE=true \
    -e DECLARE_QUEUE=true \
    -e PUBLISHER_EXCHANGE=amq.topic \
    -e SUBSCRIBER_QUEUE=cloud-event-dispatcher-queue \
    open-digital-twin/ktwin-mqtt-dispatcher:0.1
```

Build Cloud Event Dispatcher:

```sh
docker run -it --rm \
    -e SERVICE_NAME=cloud-event-dispatcher-1 \
    -e PROTOCOL=amqp \
    -e SERVER_URL=localhost \
    -e SERVER_PORT="5672" \
    -e USERNAME=rabbitmq \
    -e PASSWORD=rabbitmq \
    -e DECLARE_EXCHANGE=true \
    -e DECLARE_QUEUE=true \
    -e PUBLISHER_EXCHANGE=amq.headers \
    -e SUBSCRIBER_QUEUE=mqtt-dispatcher-queue \
    open-digital-twin/ktwin-cloud-event-dispatcher:0.1
```

## Load in Kind Development Environment

```sh
docker build -t dev.local/open-digital-twin/ktwin-mqtt-dispatcher:0.1 --build-arg SERVICE_NAME=mqtt-dispatcher .
docker build -t dev.local/open-digital-twin/ktwin-cloud-event-dispatcher:0.1 --build-arg SERVICE_NAME=cloud-event-dispatcher .
kind load docker-image dev.local/open-digital-twin/ktwin-mqtt-dispatcher:0.1
kind load docker-image dev.local/open-digital-twin/ktwin-cloud-event-dispatcher:0.1
```

## References

- [Working with Container Registry](https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-container-registry)
