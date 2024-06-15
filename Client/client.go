package Client

import (
	"Systemge/Application"
	"Systemge/Error"
	"Systemge/Message"
	"Systemge/Utilities"
	"Systemge/WebsocketServer"
	"sync"
)

type Client struct {
	name            string
	logger          *Utilities.Logger
	resolverAddress string
	randomizer      *Utilities.Randomizer

	websocketServer *WebsocketServer.Server
	application     Application.Application

	messagesWaitingForResponse map[string]chan *Message.Message // syncKey -> responseChannel

	activeBrokerConnections map[string]*brokerConnection // brokerAddress -> serverConnection
	topicResolutions        map[string]*brokerConnection // topic -> serverConnection
	mapOperationMutex       sync.Mutex

	// handleServerMessagesConcurrently is a flag that determines whether the client will handle messages concurrently
	// If this is set to true, the client will handle messages concurrently, otherwise it will handle messages sequentially
	// concurrently handling messages can be useful to improve performance but it will often require additional work for the application to be able to handle concurrency
	handleServerMessagesConcurrently      bool
	handleServerMessagesConcurrentlyMutex sync.Mutex

	stopChannel chan bool
	isStarted   bool

	stateMutex sync.Mutex
}

func New(name, topicResolutionServerAddress string, logger *Utilities.Logger, websocketServer *WebsocketServer.Server) *Client {
	return &Client{
		name:            name,
		logger:          logger,
		resolverAddress: topicResolutionServerAddress,
		randomizer:      Utilities.NewRandomizer(Utilities.GetSystemTime()),

		websocketServer: websocketServer,

		handleServerMessagesConcurrently: true,
	}
}

// sets the application that the client will use to handle messages
func (client *Client) SetApplication(application Application.Application) {
	client.stateMutex.Lock()
	defer client.stateMutex.Unlock()
	if client.isStarted {
		client.logger.Log("Cannot set application while client is started")
		return
	}
	client.application = application
}

func (client *Client) SetTopicResolutionServerAddress(address string) {
	client.resolverAddress = address
}

func (client *Client) GetApplication() Application.Application {
	return client.application
}

func (client *Client) GetWebsocketServer() *WebsocketServer.Server {
	return client.websocketServer
}

func (client *Client) GetTopicResolutionServerAddress() string {
	return client.resolverAddress
}

func (client *Client) GetLogger() *Utilities.Logger {
	return client.logger
}

func (client *Client) SetLogger(logger *Utilities.Logger) {
	client.logger = logger
}

func (client *Client) SetHandleMessagesConcurrently(handleMessagesConcurrently bool) {
	client.handleServerMessagesConcurrentlyMutex.Lock()
	defer client.handleServerMessagesConcurrentlyMutex.Unlock()
	client.handleServerMessagesConcurrently = handleMessagesConcurrently
}

func (client *Client) GetHandleMessagesConcurrently() bool {
	client.handleServerMessagesConcurrentlyMutex.Lock()
	defer client.handleServerMessagesConcurrentlyMutex.Unlock()
	return client.handleServerMessagesConcurrently
}

func (client *Client) Start() error {
	client.stateMutex.Lock()
	client.mapOperationMutex.Lock()
	if client.application == nil {
		return Error.New("Application not set", nil)
	}
	if client.resolverAddress == "" {
		return Error.New("Topic resolution server address not set", nil)
	}
	if client.isStarted {
		return Error.New("Client already connected", nil)
	}
	client.topicResolutions = make(map[string]*brokerConnection)
	client.messagesWaitingForResponse = make(map[string]chan *Message.Message)
	client.activeBrokerConnections = make(map[string]*brokerConnection)
	client.mapOperationMutex.Unlock()

	if client.websocketServer != nil {
		err := client.websocketServer.Start()
		if err != nil {
			return Error.New("Error starting websocket server", err)
		}
	}
	client.stopChannel = make(chan bool)
	topics := make([]string, 0)
	for topic := range client.application.GetSyncMessageHandlers() {
		topics = append(topics, topic)
	}
	for topic := range client.application.GetAsyncMessageHandlers() {
		topics = append(topics, topic)
	}
	for _, topic := range topics {
		serverConnection, err := client.getBrokerConnectionForTopic(topic)
		if err != nil {
			close(client.stopChannel)
			return Error.New("Error getting server connection for topic", err)
		}
		err = client.subscribeTopic(serverConnection, topic)
		if err != nil {
			close(client.stopChannel)
			return Error.New("Error subscribing to topic", err)
		}
	}
	client.isStarted = true
	client.stateMutex.Unlock()
	err := client.application.OnStart()
	if err != nil {
		client.Stop()
		return Error.New("Error in OnStart", err)
	}
	return nil
}

func (client *Client) Stop() error {
	client.stateMutex.Lock()
	if !client.isStarted {
		return Error.New("Client not connected", nil)
	}
	err := client.application.OnStop()
	if err != nil {
		return Error.New("Error in OnStop", err)
	}
	if client.websocketServer != nil {
		err := client.websocketServer.Stop()
		if err != nil {
			return Error.New("Error stopping websocket server", err)
		}
	}
	client.mapOperationMutex.Lock()
	for _, connection := range client.activeBrokerConnections {
		connection.close()
	}
	client.activeBrokerConnections = make(map[string]*brokerConnection)
	client.topicResolutions = make(map[string]*brokerConnection)
	client.messagesWaitingForResponse = make(map[string]chan *Message.Message)
	client.isStarted = false
	close(client.stopChannel)
	client.mapOperationMutex.Unlock()

	client.stateMutex.Unlock()
	return nil
}

func (server *Client) GetName() string {
	return server.name
}