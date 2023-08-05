package amqp

type Dispatcher interface {
	StartDispatcher() error
	Listen() error
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
}

type dispatcher struct {
	config DispatcherConfig
}

func (*dispatcher) StartDispatcher() error {
	return nil
}

func (*dispatcher) Listen() error {
	return nil
}
