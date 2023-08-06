package amqp

import (
	"fmt"

	"github.com/agwermann/ktwin-dispatcher/pkg/config"
)

type Dispatcher interface {
	Start() error
	CloseDispatcher()
	Listen(callback DispatcherCallback) error
}

func NewDispatcher(config DispatcherConfig) Dispatcher {
	return &dispatcher{
		config: config,
	}
}

type DispatcherConfig struct {
	Protocol   string
	ServerUrl  string
	ServerPort int
	Username   string
	Password   string

	SubscriberQueue string
	PublisherTopic  string
	ServiceName     string

	ExchangeType string
}

type DispatcherCallbackParams struct {
	Body            []byte
	Headers         map[string]interface{}
	RoutingKey      string
	PublisherClient AMQPClient
}

type DispatcherCallback func(params DispatcherCallbackParams)

type dispatcher struct {
	config           DispatcherConfig
	subscriberClient AMQPClient
	publisherClient  AMQPClient
}

func (d *dispatcher) Start() error {
	d.subscriberClient = NewAMQPClient()
	err := d.subscriberClient.Connect(d.config.ServerUrl, d.config.ServerPort, d.config.Protocol, d.config.Username, d.config.Password)

	if err != nil {
		fmt.Println("Error when connecting subscriber to broker")
		return err
	}

	d.publisherClient = NewAMQPClient()
	err = d.publisherClient.Connect(d.config.ServerUrl, d.config.ServerPort, d.config.Protocol, d.config.Username, d.config.Password)

	if err != nil {
		fmt.Println("Error when connecting publisher to broker")
		return err
	}

	err = d.publisherClient.DeclareExchange(d.config.PublisherTopic, d.config.ExchangeType, ExchangeOptions{
		Durable:    true,
		AutoDelete: false,
		Internal:   false,
		NoWait:     false,
	})

	if err != nil {
		fmt.Println("Error when declaring publisher exchange")
		return err
	}

	err = d.subscriberClient.DeclareQueue(d.config.SubscriberQueue, QueueOptions{
		Durable:    true,
		AutoDelete: false,
		Exclusive:  false,
		NoWait:     false,
	})

	if err != nil {
		fmt.Println("Error when declaring subscriber queue")
		return err
	}

	return nil
}

func (d *dispatcher) CloseDispatcher() {
	defer d.subscriberClient.Close()
	defer d.publisherClient.Close()
}

func (d *dispatcher) Listen(callback DispatcherCallback) error {
	err := d.subscriberClient.Consume(d.config.SubscriberQueue, d.config.ServiceName, ConsumerClientOptions{
		AutoAck:   true,
		Exclusive: false,
		NoLocal:   false,
		NoWait:    false,
	}, func(body []byte, headers map[string]interface{}, routingKey string) {

		if config.GetEnv("APP_ENV") == "local" {
			fmt.Printf("Received message: %s %s\n", routingKey, body)
		}

		params := DispatcherCallbackParams{
			Body:            body,
			Headers:         headers,
			RoutingKey:      routingKey,
			PublisherClient: d.publisherClient,
		}
		callback(params)
	})

	if err != nil {
		fmt.Println("Error when consuming from subscriber queue")
		return err
	}

	return nil
}
