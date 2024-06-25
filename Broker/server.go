package Broker

import (
	"Systemge/Error"
	"Systemge/Utilities"
	"crypto/tls"
	"net"
	"sync"
)

type Server struct {
	name   string
	logger *Utilities.Logger

	syncTopics  map[string]bool
	asyncTopics map[string]bool

	clientSubscriptions map[string]map[string]*clientConnection // topic -> [clientName-> client]
	clientConnections   map[string]*clientConnection            // clientName -> Client
	openSyncRequests    map[string]*syncRequest                 // syncKey -> request

	brokerTlsCertPath string
	brokerTlsKeyPath  string
	brokerPort        string
	tlsBrokerListener net.Listener

	configTlsCertPath string
	configTlsKeyPath  string
	configPort        string
	tlsConfigListener net.Listener

	isStarted bool

	operationMutex sync.Mutex
	stateMutex     sync.Mutex
}

func New(name, brokerPort, brokerTlsCertPath, brokerTlsKeyPath, configPort, configTlsCertPath, configTlsKeyPath string, logger *Utilities.Logger) *Server {
	return &Server{
		name:   name,
		logger: logger,

		syncTopics: map[string]bool{
			"subscribe":   true,
			"unsubscribe": true,
			"consume":     true,
		},
		asyncTopics: map[string]bool{
			"heartbeat": true,
		},

		clientSubscriptions: map[string]map[string]*clientConnection{},
		clientConnections:   map[string]*clientConnection{},
		openSyncRequests:    map[string]*syncRequest{},

		brokerTlsCertPath: brokerTlsCertPath,
		brokerTlsKeyPath:  brokerTlsKeyPath,
		brokerPort:        brokerPort,

		configTlsCertPath: configTlsCertPath,
		configTlsKeyPath:  configTlsKeyPath,
		configPort:        configPort,
	}
}

func (server *Server) Start() error {
	server.stateMutex.Lock()
	defer server.stateMutex.Unlock()
	if server.isStarted {
		return Error.New("Server already started", nil)
	}
	brokerCert, err := tls.LoadX509KeyPair(server.brokerTlsCertPath, server.brokerTlsKeyPath)
	if err != nil {
		return Error.New("Failed to load TLS certificate: ", err)
	}
	brokerListener, err := tls.Listen("tcp", server.brokerPort, &tls.Config{
		MinVersion:   tls.VersionTLS12,
		Certificates: []tls.Certificate{brokerCert},
	})
	if err != nil {
		return Error.New("Failed to start server: ", err)
	}
	configCert, err := tls.LoadX509KeyPair(server.configTlsCertPath, server.configTlsKeyPath)
	if err != nil {
		return Error.New("Failed to load TLS certificate: ", err)
	}
	configListener, err := tls.Listen("tcp", server.configPort, &tls.Config{
		MinVersion:   tls.VersionTLS12,
		Certificates: []tls.Certificate{configCert},
	})
	if err != nil {
		return Error.New("Failed to start server: ", err)
	}
	server.tlsBrokerListener = brokerListener
	server.tlsConfigListener = configListener
	server.isStarted = true
	go server.handleClientConnections()
	go server.handleConfigConnections()
	return nil
}

func (server *Server) GetName() string {
	return server.name
}

func (server *Server) Stop() error {
	server.stateMutex.Lock()
	defer server.stateMutex.Unlock()
	if !server.isStarted {
		return Error.New("Server is not started", nil)
	}
	server.tlsBrokerListener.Close()
	server.tlsConfigListener.Close()
	server.disconnectAllClientConnections()
	server.isStarted = false
	return nil
}

func (server *Server) IsStarted() bool {
	server.stateMutex.Lock()
	defer server.stateMutex.Unlock()
	return server.isStarted
}
