package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	amqp "github.com/agwermann/ktwin-dispatcher/pkg/amqp"
	"github.com/agwermann/ktwin-dispatcher/pkg/config"
)

func main() {
	ctx := context.Background()
	config.LoadEnv()

	protocol := config.GetEnv("PROTOCOL")
	serverUrl := config.GetEnv("SERVER_URL")
	serverPort := config.GetEnvInt("SERVER_PORT")
	username := config.GetEnv("USERNAME")
	password := config.GetEnv("PASSWORD")

	subscriberQueue := config.GetEnv("SUBSCRIBER_QUEUE")
	publisherTopic := config.GetEnv("PUBLISHER_TOPIC")
	serviceName := config.GetEnv("SERVICE_NAME")

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
		newRoutingKey, err := buildRoutingKeyWithHeaders(params.Headers)

		if err == nil {
			message := amqp.Message{
				ContentType: "application/json",
				Body:        params.Body,
			}

			err = params.PublisherClient.Publish(ctx, message, publisherTopic, newRoutingKey, amqp.PublisherClientOptions{
				Mandatory: false,
				Immediate: false,
			})

			if err != nil {
				fmt.Println(err)
			}
		} else {
			fmt.Printf("Routing key has invalid format: %s %s, dropping the message", params.Headers["type"], params.Headers["source"])
		}

	})

	<-forever

}

func buildRoutingKeyWithHeaders(headers map[string]interface{}) (string, error) {
	headerType, isTypeValid := headers["type"].(string)
	sourceType, isSourceValid := headers["source"].(string)

	if isTypeValid && isSourceValid {
		return strings.Join([]string{headerType, sourceType}, "."), nil
	}

	return "", errors.New("Invalid routing key")
}
