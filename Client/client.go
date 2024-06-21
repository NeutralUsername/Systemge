package Client

import (
	"Systemge/Application"
	"Systemge/HTTPServer"
	"Systemge/Message"
	"Systemge/Utilities"
	"Systemge/WebsocketServer"
	"sync"
)

type Client struct {
	name            string
	logger          *Utilities.Logger
	resolverAddress string

	httpServer      *HTTPServer.Server
	websocketServer *WebsocketServer.Server
	application     Application.Application

	messagesWaitingForResponse map[string]chan *Message.Message // syncKey -> responseChannel

	activeBrokerConnections map[string]*brokerConnection // brokerAddress -> serverConnection
	topicResolutions        map[string]*brokerConnection // topic -> serverConnection
	mapOperationMutex       sync.Mutex

	// handleMessagesConcurrently is a flag that determines whether the client will handle messages by brokers concurrently
	// If this is set to true, the client will handle messages concurrently, otherwise it will handle messages sequentially
	// concurrently handling messages can be useful to improve performance but it will often require additional work for the application to be able to handle concurrency
	handleMessagesConcurrently      bool
	handleMessagesConcurrentlyMutex sync.Mutex

	stopChannel chan bool
	isStarted   bool

	stateMutex sync.Mutex
}

func New(name, topicResolutionServerAddress string, logger *Utilities.Logger) *Client {
	return &Client{
		name:            name,
		logger:          logger,
		resolverAddress: topicResolutionServerAddress,

		messagesWaitingForResponse: make(map[string]chan *Message.Message),
		topicResolutions:           make(map[string]*brokerConnection),
		activeBrokerConnections:    make(map[string]*brokerConnection),

		httpServer:      nil,
		websocketServer: nil,

		handleMessagesConcurrently: true,
	}
}

func (client *Client) SetApplication(application Application.Application) {
	client.application = application
}

func (client *Client) SetWebsocketServer(websocketServer *WebsocketServer.Server) {
	client.websocketServer = websocketServer
}

func (client *Client) SetHTTPServer(httpServer *HTTPServer.Server) {
	client.httpServer = httpServer
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

func (client *Client) GetHTTPServer() *HTTPServer.Server {
	return client.httpServer
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
	client.handleMessagesConcurrentlyMutex.Lock()
	defer client.handleMessagesConcurrentlyMutex.Unlock()
	client.handleMessagesConcurrently = handleMessagesConcurrently
}

func (client *Client) GetHandleMessagesConcurrently() bool {
	client.handleMessagesConcurrentlyMutex.Lock()
	defer client.handleMessagesConcurrentlyMutex.Unlock()
	return client.handleMessagesConcurrently
}

func (client *Client) Start() error {
	err := func() error {
		client.stateMutex.Lock()
		defer client.stateMutex.Unlock()
		if client.application == nil {
			return Utilities.NewError("Application not set", nil)
		}
		if client.resolverAddress == "" {
			return Utilities.NewError("Topic resolution server address not set", nil)
		}
		if client.isStarted {
			return Utilities.NewError("Client already connected", nil)
		}
		if client.websocketServer != nil {
			err := client.websocketServer.Start()
			if err != nil {
				return Utilities.NewError("Error starting websocket server", err)
			}
		}
		if client.httpServer != nil {
			err := client.httpServer.Start()
			if err != nil {
				return Utilities.NewError("Error starting http server", err)
			}
		}
		client.stopChannel = make(chan bool)
		topics := make([]string, 0)

		client.mapOperationMutex.Lock()
		for topic := range client.application.GetSyncMessageHandlers() {
			topics = append(topics, topic)
		}
		for topic := range client.application.GetAsyncMessageHandlers() {
			topics = append(topics, topic)
		}
		client.mapOperationMutex.Unlock()

		for _, topic := range topics {
			serverConnection, err := client.getBrokerConnectionForTopic(topic)
			if err != nil {
				client.removeAllBrokerConnections()
				close(client.stopChannel)
				return Utilities.NewError("Error getting server connection for topic", err)
			}
			err = client.subscribeTopic(serverConnection, topic)
			if err != nil {
				client.removeAllBrokerConnections()
				close(client.stopChannel)
				return Utilities.NewError("Error subscribing to topic", err)
			}
		}
		client.isStarted = true
		return nil
	}()
	if err != nil {
		return Utilities.NewError("Error starting client", err)
	}
	err = client.application.OnStart()
	if err != nil {
		client.Stop()
		return Utilities.NewError("Error in OnStart", err)
	}
	return nil
}

func (client *Client) Stop() error {
	client.stateMutex.Lock()
	defer client.stateMutex.Unlock()
	if !client.isStarted {
		return Utilities.NewError("Client not started", nil)
	}
	err := client.application.OnStop()
	if err != nil {
		return Utilities.NewError("Error in OnStop", err)
	}
	if client.websocketServer != nil {
		err := client.websocketServer.Stop()
		if err != nil {
			return Utilities.NewError("Error stopping websocket server", err)
		}
	}
	if client.httpServer != nil {
		err := client.httpServer.Stop()
		if err != nil {
			return Utilities.NewError("Error stopping http server", err)
		}
	}
	client.isStarted = false
	client.removeAllBrokerConnections()
	close(client.stopChannel)
	return nil
}

func (client *Client) GetName() string {
	return client.name
}

func (client *Client) IsStarted() bool {
	client.stateMutex.Lock()
	defer client.stateMutex.Unlock()
	return client.isStarted
}
