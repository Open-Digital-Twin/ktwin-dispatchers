package amqp

import (
	"context"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type ConsumerClientOptions struct {
	AutoAck   bool
	Exclusive bool
	NoLocal   bool
	NoWait    bool
}

type PublisherClientOptions struct {
	Mandatory bool
	Immediate bool
}

type ExchangeOptions struct {
	Durable    bool
	AutoDelete bool
	Internal   bool
	NoWait     bool
}

type QueueOptions struct {
	Durable    bool
	AutoDelete bool
	Exclusive  bool
	NoWait     bool
}

type ConsumerCallback func(body []byte, headers map[string]interface{}, routingKey string)

type AMQPClient interface {
	Connect(serverUrl string, port int, protocol string, username string, password string) error
	Close() error
	Consume(queue string, consumer string, options ConsumerClientOptions, callback ConsumerCallback) error
	// kind: direct, fanout, topic, headers
	DeclareExchange(exchange string, kind string, options ExchangeOptions) error
	DeclareQueue(queue string, options QueueOptions) error
	Publish(ctx context.Context, message Message, exchange string, key string, options PublisherClientOptions) error
}

type amqpClient struct {
	connection *amqp.Connection
	channel    *amqp.Channel
}

func NewAMQPClient() AMQPClient {
	return &amqpClient{}
}

func (c *amqpClient) Connect(serverUrl string, port int, protocol string, username string, password string) error {
	var err error
	var url string
	if username != "" && password != "" {
		url = fmt.Sprintf("%s://%s:%s@%s:%d", protocol, username, password, serverUrl, port)
	} else {
		url = fmt.Sprintf("%s://%s:%d", protocol, serverUrl, port)
	}
	c.connection, err = amqp.Dial(url)

	if err != nil {
		return err
	}

	c.channel, err = c.connection.Channel()
	if err != nil {
		panic(err)
	}

	return err
}

func (c *amqpClient) Close() error {
	defer c.connection.Close()
	defer c.channel.Close()
	return nil
}

func (c *amqpClient) Consume(queue string, consumer string, options ConsumerClientOptions, callback ConsumerCallback) error {

	messages, err := c.channel.Consume(queue, consumer, options.AutoAck, options.Exclusive, options.NoLocal, options.NoWait, nil)

	if err != nil {
		log.Println(err)
		return err
	}

	forever := make(chan bool)

	go func() {
		for message := range messages {
			callback(message.Body, message.Headers, message.RoutingKey)
		}
	}()

	<-forever

	return nil
}

func (c *amqpClient) DeclareExchange(exchange string, kind string, options ExchangeOptions) error {
	return c.channel.ExchangeDeclare(
		exchange,
		kind,
		options.Durable,
		options.AutoDelete,
		options.Internal,
		options.NoWait,
		nil,
	)
}

type Message struct {
	Headers     map[string]interface{}
	ContentType string
	Body        []byte
	MessageId   string
}

func (c *amqpClient) Publish(ctx context.Context, message Message, exchange string, key string, options PublisherClientOptions) error {
	amqpMessage := amqp.Publishing{
		Headers:      message.Headers,
		ContentType:  message.ContentType,
		DeliveryMode: amqp.Persistent,
		Body:         message.Body,
	}
	return c.channel.PublishWithContext(ctx, exchange, key, options.Mandatory, options.Immediate, amqpMessage)
}

func (c *amqpClient) DeclareQueue(queue string, options QueueOptions) error {
	properties := amqp.NewConnectionProperties()
	properties["x-queue-type"] = "quorum"
	_, err := c.channel.QueueDeclare(queue, options.Durable, options.AutoDelete, options.Exclusive, options.NoWait, properties)
	return err
}
