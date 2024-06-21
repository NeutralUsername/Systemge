package Module

import (
	"Systemge/Application"
	"Systemge/Client"
	"Systemge/HTTPServer"
	"Systemge/Utilities"
	"Systemge/WebsocketServer"
)

type NewApplicationFunc func(*Client.Client, []string) (Application.Application, error)
type NewCompositeApplicationWebsocketFunc func(*Client.Client, []string) (Application.CompositeApplicationWebsocket, error)
type NewCompositeApplicationHTTPFunc func(*Client.Client, []string) (Application.CompositeApplicationHTTP, error)
type NewCompositeApplicationtWebsocketHTTPFunc func(*Client.Client, []string) (Application.CompositeApplicationWebsocketHTTP, error)

func NewClient(name string, resolverAddress string, loggerPath string, newApplicationFunc NewApplicationFunc, args []string) *Client.Client {
	client := Client.New(name, resolverAddress, Utilities.NewLogger(loggerPath))
	application, err := newApplicationFunc(client, args)
	if err != nil {
		panic(err)
	}
	client.SetApplication(application)
	return client
}

func NewCompositeClientHTTP(name string, resolverAddress string, loggerPath string, httpPort string, httpTlsCert string, httpTlsKey string, newCompositeApplicationHTTPFunc NewCompositeApplicationHTTPFunc, args []string) *Client.Client {
	client := Client.New(name, resolverAddress, Utilities.NewLogger(loggerPath))
	application, err := newCompositeApplicationHTTPFunc(client, args)
	if err != nil {
		panic(err)
	}
	httpServer := HTTPServer.New(httpPort, name+"HTTP", httpTlsCert, httpTlsKey, Utilities.NewLogger(loggerPath), application)
	client.SetApplication(application)
	client.SetHTTPServer(httpServer)
	return client
}

func NewCompositeClientWebsocket(name string, resolverAddress string, loggerPath string, websocketPattern string, websocketPort string, websocketTlsCert string, websocketTlsKey string, newCompositeApplicationWebsocketFunc NewCompositeApplicationWebsocketFunc, args []string) *Client.Client {
	client := Client.New(name, resolverAddress, Utilities.NewLogger(loggerPath))
	application, err := newCompositeApplicationWebsocketFunc(client, args)
	if err != nil {
		panic(err)
	}
	websocketServer := WebsocketServer.New(name, Utilities.NewLogger(loggerPath), application)
	websocketServer.SetHTTPServer(HTTPServer.New(websocketPort, name+"HTTP", websocketTlsCert, websocketTlsKey, Utilities.NewLogger(loggerPath), WebsocketServer.NewHandshakeApplication(websocketPattern, websocketServer)))
	client.SetApplication(application)
	client.SetWebsocketServer(websocketServer)
	return client
}

func NewCompositeClientWebsocketHTTP(name string, resolverAddress string, loggerPath string, websocketPattern string, websocketPort string, websocketTlsCert string, websocketTlsKey string, httpPort string, httpTlsCert string, httpTlsKey string, newCompositeApplicationtWebsocketHTTPFunc NewCompositeApplicationtWebsocketHTTPFunc, args []string) *Client.Client {
	client := Client.New(name, resolverAddress, Utilities.NewLogger(loggerPath))
	application, err := newCompositeApplicationtWebsocketHTTPFunc(client, args)
	if err != nil {
		panic(err)
	}
	websocketServer := WebsocketServer.New(name, Utilities.NewLogger(loggerPath), application)
	websocketServer.SetHTTPServer(HTTPServer.New(websocketPort, name+"HTTP", websocketTlsCert, websocketTlsKey, Utilities.NewLogger(loggerPath), WebsocketServer.NewHandshakeApplication(websocketPattern, websocketServer)))
	httpServer := HTTPServer.New(httpPort, name+"HTTP", httpTlsCert, httpTlsKey, Utilities.NewLogger(loggerPath), application)
	client.SetApplication(application)
	client.SetWebsocketServer(websocketServer)
	client.SetHTTPServer(httpServer)
	return client
}
