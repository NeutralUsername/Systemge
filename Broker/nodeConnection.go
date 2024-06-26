package Broker

import (
	"Systemge/Error"
	"Systemge/Message"
	"Systemge/Utilities"
	"net"
	"sync"
	"time"
)

type nodeConnection struct {
	name         string
	netConn      net.Conn
	messageQueue chan *Message.Message
	watchdog     *time.Timer

	subscribedTopics   map[string]bool
	deliverImmediately bool

	watchdogMutex sync.Mutex
	sendMutex     sync.Mutex
	receiveMutex  sync.Mutex

	stopChannel chan bool
}

func newNodeConnection(name string, netConn net.Conn) *nodeConnection {
	return &nodeConnection{
		name:               name,
		netConn:            netConn,
		messageQueue:       make(chan *Message.Message, NODE_MESSAGE_QUEUE_SIZE),
		watchdog:           nil,
		subscribedTopics:   map[string]bool{},
		deliverImmediately: DELIVER_IMMEDIATELY_DEFAULT,
		stopChannel:        make(chan bool),
	}
}

func (nodeConnection *nodeConnection) send(message *Message.Message) error {
	nodeConnection.sendMutex.Lock()
	defer nodeConnection.sendMutex.Unlock()
	return Utilities.TcpSend(nodeConnection.netConn, message.Serialize(), DEFAULT_TCP_TIMEOUT)
}

func (nodeConnection *nodeConnection) receive() ([]byte, error) {
	nodeConnection.receiveMutex.Lock()
	defer nodeConnection.receiveMutex.Unlock()
	messageBytes, err := Utilities.TcpReceive(nodeConnection.netConn, 0)
	if err != nil {
		return nil, err
	}
	return messageBytes, nil
}

func (nodeConnection *nodeConnection) resetWatchdog() error {
	nodeConnection.watchdogMutex.Lock()
	defer nodeConnection.watchdogMutex.Unlock()
	if nodeConnection.watchdog == nil {
		return Error.New("Watchdog is not set for node \""+nodeConnection.name+"\"", nil)
	}
	nodeConnection.watchdog.Reset((WATCHDOG_TIMEOUT))
	return nil
}

func (nodeConnection *nodeConnection) disconnect() error {
	nodeConnection.watchdogMutex.Lock()
	if nodeConnection.watchdog == nil {
		return Error.New("Watchdog is not set for node \""+nodeConnection.name+"\"", nil)
	}
	nodeConnection.watchdog.Reset(0)
	<-nodeConnection.stopChannel
	nodeConnection.watchdogMutex.Unlock()
	return nil
}

func (broker *Broker) addNodeConnection(nodeConnection *nodeConnection) error {
	broker.operationMutex.Lock()
	defer broker.operationMutex.Unlock()
	if broker.nodeConnections[nodeConnection.name] != nil {
		return Error.New("node with name \""+nodeConnection.name+"\" already exists", nil)
	}
	broker.nodeConnections[nodeConnection.name] = nodeConnection
	nodeConnection.watchdog = time.AfterFunc(WATCHDOG_TIMEOUT, func() {
		watchdog := nodeConnection.watchdog
		nodeConnection.watchdog = nil
		watchdog.Stop()
		nodeConnection.netConn.Close()
		err := broker.removeNodeConnection(nodeConnection)
		if err != nil {
			broker.logger.Log(Error.New("Error removing node \""+nodeConnection.name+"\"", err).Error())
		}
		close(nodeConnection.stopChannel)
	})
	return nil
}

func (broker *Broker) removeNodeConnection(nodeConnection *nodeConnection) error {
	broker.operationMutex.Lock()
	defer broker.operationMutex.Unlock()
	if broker.nodeConnections[nodeConnection.name] == nil {
		return Error.New("Subscriber with name \""+nodeConnection.name+"\" does not exist", nil)
	}
	for messageType := range nodeConnection.subscribedTopics {
		delete(broker.nodeSubscriptions[messageType], nodeConnection.name)
	}
	delete(broker.nodeConnections, nodeConnection.name)
	return nil
}

func (broker *Broker) disconnectAllNodeConnections() {
	NodesToDisconnect := make([]*nodeConnection, 0)
	broker.operationMutex.Lock()
	for _, nodeConnection := range broker.nodeConnections {
		NodesToDisconnect = append(NodesToDisconnect, nodeConnection)
	}
	broker.operationMutex.Unlock()
	for _, nodeConnection := range NodesToDisconnect {
		nodeConnection.disconnect()
	}
}
