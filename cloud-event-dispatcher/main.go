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

	dispatcherConfig := amqp.DispatcherConfig{
		Protocol:        protocol,
		ServerUrl:       serverUrl,
		ServerPort:      serverPort,
		Username:        username,
		Password:        password,
		SubscriberQueue: subscriberQueue,
		PublisherTopic:  publisherTopic,
		ServiceName:     serviceName,
		ExchangeType:    "topic",
	}

	dispatcher := amqp.NewDispatcher(dispatcherConfig)

	err := dispatcher.Start()

	if err != nil {
		fmt.Println(err)
		panic("Error when starting dispatcher")
	}

	defer dispatcher.CloseDispatcher()

	forever := make(chan bool)

	dispatcher.Listen(func(params amqp.DispatcherCallbackParams) {
		// Convert the `type` header to routing key sent to MQTT topic exchange
		newRoutingKey := params.Headers["type"].(string)
		message := amqp.Message{
			ContentType: "application/json",
			Body:        params.Body,
		}

		err := params.PublisherClient.Publish(ctx, message, publisherTopic, newRoutingKey, amqp.PublisherClientOptions{
			Mandatory: false,
			Immediate: false,
		})

		if err != nil {
			fmt.Println(err)
		}
	})

	<-forever

}
