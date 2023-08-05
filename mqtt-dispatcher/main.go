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
		ExchangeType:    "headers",
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
		fmt.Printf("Message received: %s\n", string(params.Body))
		fmt.Printf("Headers received: %s\n", params.Headers)
		fmt.Printf("Routing Key received: %s\n", params.RoutingKey)

		headers := make(map[string]interface{}, 1)
		headers["type"] = params.RoutingKey

		message := amqp.Message{
			Headers:     headers,
			ContentType: "application/json",
			Body:        params.Body,
		}

		err := params.PublisherClient.Publish(ctx, message, publisherTopic, "", amqp.PublisherClientOptions{
			Mandatory: false,
			Immediate: false,
		})

		if err != nil {
			fmt.Println(err)
		}
	})

	<-forever

}
