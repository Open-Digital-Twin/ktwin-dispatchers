package main

import (
	"context"
	"fmt"

	amqp "github.com/agwermann/ktwin-dispatcher/pkg/amqp"
	"github.com/agwermann/ktwin-dispatcher/pkg/config"
)

func main() {
	ctx := context.Background()
	config.LoadEnv()

	protocol := config.GetEnv("protocol")
	serverUrl := config.GetEnv("serverUrl")
	serverPort := config.GetEnvInt("serverPort")
	username := config.GetEnv("username")
	password := config.GetEnv("password")

	subscriberQueue := config.GetEnv("subscriberQueue")
	publisherTopic := config.GetEnv("publisherTopic")
	serviceName := config.GetEnv("serviceName")

	amqpSubscriberClient := amqp.NewAMQPClient()
	err := amqpSubscriberClient.Connect(serverUrl, serverPort, protocol, username, password)
	defer amqpSubscriberClient.Close()

	if err != nil {
		fmt.Println(err)
		panic("Error while connecting")
	}

	amqpPublisherClient := amqp.NewAMQPClient()
	err = amqpPublisherClient.Connect(serverUrl, serverPort, protocol, username, password)
	defer amqpPublisherClient.Close()

	if err != nil {
		fmt.Println(err)
		panic("Error while connecting")
	}

	err = amqpPublisherClient.DeclareExchange(publisherTopic, "headers", amqp.ExchangeOptions{
		Durable:    true,
		AutoDelete: false,
		Internal:   false,
		NoWait:     false,
	})

	if err != nil {
		fmt.Println(err)
		panic("Error while connecting")
	}

	err = amqpSubscriberClient.DeclareQueue(subscriberQueue, amqp.QueueOptions{
		Durable:    true,
		AutoDelete: false,
		Exclusive:  false,
		NoWait:     false,
	})

	if err != nil {
		panic("Error while connecting")
	}

	forever := make(chan bool)

	amqpSubscriberClient.Consume(subscriberQueue, serviceName, amqp.ConsumerClientOptions{
		AutoAck:   true,
		Exclusive: false,
		NoLocal:   false,
		NoWait:    false,
	}, func(body []byte, headers map[string]interface{}, routingKey string) {
		fmt.Printf("Message received: %s\n", string(body))
		fmt.Printf("Headers received: %s\n", headers)
		fmt.Printf("Routing Key received: %s\n", routingKey)

		headers = make(map[string]interface{}, 1)
		headers["type"] = routingKey

		message := amqp.Message{
			Headers:     headers,
			ContentType: "application/json",
			Body:        body,
		}

		err := amqpPublisherClient.Publish(ctx, message, publisherTopic, "", amqp.PublisherClientOptions{
			Mandatory: false,
			Immediate: false,
		})

		if err != nil {
			fmt.Println(err)
		}
	})

	<-forever

}
