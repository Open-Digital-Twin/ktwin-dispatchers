package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	amqp "github.com/Open-Digital-Twin/ktwin-dispatchers/pkg/amqp"
	"github.com/Open-Digital-Twin/ktwin-dispatchers/pkg/config"
	"github.com/Open-Digital-Twin/ktwin-dispatchers/pkg/metrics"
	cloudEvents "github.com/cloudevents/sdk-go/v2"
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

	declareExchange := config.GetEnvBool("DECLARE_EXCHANGE")
	declareQueue := config.GetEnvBool("DECLARE_QUEUE")
	subscriberQueue := config.GetEnv("SUBSCRIBER_QUEUE")
	publisherExchange := config.GetEnv("PUBLISHER_EXCHANGE")
	serviceName := config.GetEnv("SERVICE_NAME")

	dispatcherConfig := amqp.DispatcherConfig{
		Protocol:          protocol,
		ServerUrl:         serverUrl,
		ServerPort:        serverPort,
		Username:          username,
		Password:          password,
		DeclareExchange:   declareExchange,
		DeclareQueue:      declareQueue,
		SubscriberQueue:   subscriberQueue,
		PublisherExchange: publisherExchange,
		ServiceName:       serviceName,
		ExchangeType:      "headers",
	}

	metricsServer := metrics.NewMetricsServer()
	go metricsServer.ServeMetrics()

	dispatcher := amqp.NewDispatcher(dispatcherConfig)

	err := dispatcher.Start()

	if err != nil {
		fmt.Println(err)
		panic("Error when starting dispatcher")
	}

	defer dispatcher.CloseDispatcher()

	forever := make(chan bool)

	dispatcher.Listen(func(params amqp.DispatcherCallbackParams) {
		start := time.Now()

		// Convert the routing key to type header used by Cloud Events header exchange
		headers := make(map[string]interface{}, 10)

		headerType, headerSource, err := splitRoutingKeyInCloudEvents(params.RoutingKey)

		if err == nil {
			messageId := uuid.NewString()
			headers["content-type"] = "application/json"
			headers["id"] = messageId
			headers["time"] = cloudEvents.Timestamp{Time: time.Now().UTC()}.String()
			headers["specversion"] = "1.0"
			headers["type"] = headerType
			headers["source"] = headerSource

			message := amqp.Message{
				Headers:     headers,
				ContentType: "application/json",
				Body:        params.Body,
				MessageId:   messageId,
			}

			err = params.PublisherClient.Publish(ctx, message, publisherExchange, "", amqp.PublisherClientOptions{
				Mandatory: false,
				Immediate: false,
			})

			duration := time.Since(start)

			if err != nil {
				fmt.Println(err)
				metricsServer.IncrementEventCount(headerType, http.StatusInternalServerError)
				metricsServer.ObserveRequestDuration(headerType, http.StatusInternalServerError, float64(duration.Milliseconds()))
			} else {
				metricsServer.IncrementEventCount(headerType, http.StatusAccepted)
				metricsServer.ObserveRequestDuration(headerType, http.StatusAccepted, float64(duration.Milliseconds()))
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
