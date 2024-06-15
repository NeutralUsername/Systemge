package MessageBrokerClient

import (
	"Systemge/Error"
	"Systemge/Message"
	"Systemge/ResolverServer"
	"Systemge/TCP"
	"Systemge/Utilities"
	"net"
	"sync"
)

type serverConnection struct {
	netConn    net.Conn
	resolution *ResolverServer.Resolution
	logger     *Utilities.Logger

	topics            map[string]bool
	mapOperationMutex sync.Mutex

	sendMutex    sync.Mutex
	receiveMutex sync.Mutex
}

func newServerConnection(netConn net.Conn, resolution *ResolverServer.Resolution, logger *Utilities.Logger) *serverConnection {
	return &serverConnection{
		netConn:    netConn,
		resolution: resolution,
		logger:     logger,

		topics: make(map[string]bool),
	}
}

func (serverConnection *serverConnection) send(message *Message.Message) error {
	if serverConnection == nil {
		return Error.New("Server connection is nil", nil)
	}
	serverConnection.sendMutex.Lock()
	defer serverConnection.sendMutex.Unlock()
	if serverConnection.netConn == nil {
		return Error.New("Connection is closed", nil)
	}
	err := TCP.Send(serverConnection.netConn, message.Serialize(), DEFAULT_TCP_TIMEOUT)
	if err != nil {
		return Error.New("Error sending message", err)
	}
	return nil
}

func (serverConnection *serverConnection) receive() ([]byte, error) {
	if serverConnection == nil {
		return nil, Error.New("Server connection is nil", nil)
	}
	serverConnection.receiveMutex.Lock()
	defer serverConnection.receiveMutex.Unlock()
	if serverConnection.netConn == nil {
		return nil, Error.New("Connection is closed", nil)
	}
	messageBytes, err := TCP.Receive(serverConnection.netConn, 0)
	if err != nil {
		return nil, Error.New("Error receiving message", err)
	}
	return messageBytes, nil
}

func (serverConnection *serverConnection) close() error {
	if serverConnection == nil {
		return Error.New("Server connection is nil", nil)
	}
	if serverConnection.netConn == nil {
		return Error.New("Connection is already closed", nil)
	}
	serverConnection.netConn.Close()
	serverConnection.netConn = nil
	return nil
}

func (serverConnection *serverConnection) addTopic(topic string) error {
	serverConnection.mapOperationMutex.Lock()
	defer serverConnection.mapOperationMutex.Unlock()
	if serverConnection.topics[topic] {
		return Error.New("Topic already exists", nil)
	}
	serverConnection.topics[topic] = true
	return nil
}
