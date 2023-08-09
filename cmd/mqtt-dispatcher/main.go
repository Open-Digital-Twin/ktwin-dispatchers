package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	amqp "github.com/agwermann/ktwin-dispatcher/pkg/amqp"
	"github.com/agwermann/ktwin-dispatcher/pkg/config"
	"github.com/google/uuid"
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
		// Convert the routing key to type header used by Cloud Events header exchange
		headers := make(map[string]interface{}, 1)

		headerType, headerSource, err := splitRoutingKeyInCloudEvents(params.RoutingKey)

		if err == nil {
			headers["id"] = uuid.NewString()
			headers["specversion"] = "1.0"
			headers["type"] = headerType
			headers["source"] = headerSource

			message := amqp.Message{
				Headers:     headers,
				ContentType: "application/json",
				Body:        params.Body,
			}

			err = params.PublisherClient.Publish(ctx, message, publisherTopic, "", amqp.PublisherClientOptions{
				Mandatory: false,
				Immediate: false,
			})

			if err != nil {
				fmt.Println(err)
			}
		} else {
			fmt.Printf("Invalid routing key format %s, message will be dropped\n", params.RoutingKey)
		}

	})

	<-forever

}

// Split the routing key (e.g. MQTT topic) into CloudEvent headers
// RoutingKey format: ktwin.real.twin-interface.twin-instance
// Return type, source cloud events
func splitRoutingKeyInCloudEvents(routingKey string) (string, string, error) {
	splitRK := strings.Split(routingKey, ".")

	if len(splitRK) < 4 {
		return "", "", errors.New("Invalid format")
	}

	headerType := strings.Join(splitRK[0:3], ".")
	sourceType := splitRK[3]

	return headerType, sourceType, nil
}
