# KTwin Dispatcher Repository

This repository contains the MQTT and Cloud Event Dispatchers used to route MQTT messages and AMQP messages within RabbitMQ cluster running under Knative Broker abstraction.

KTwin uses RabbitMQ 3.12 which provides Native MQTT functionalities which allows serving millions of MQTT devices. RabbitMQ offers protocol interoperability between AMQP and MQTT. More details about how RabbitMQ Native MQTT [here](https://blog.rabbitmq.com/posts/2023/03/native-mqtt/).

The Cloud Event Dispatcher is responsible to forward Cloud Event messages from Virtual components generated within KTwin cluster to real devices connected to the cluster using MQTT. The MQTT Dispatcher forwards MQTT messages from real components to virtual services.
