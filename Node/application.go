package Node

import (
	"Systemge/Config"
	"Systemge/Message"
	"net/http"
)

type Application interface {
	GetAsyncMessageHandlers() map[string]AsyncMessageHandler
	GetSyncMessageHandlers() map[string]SyncMessageHandler
	GetCustomCommandHandlers() map[string]CustomCommandHandler
	OnStart(*Node) error
	OnStop(*Node) error
	GetApplicationConfig() Config.Application
}
type AsyncMessageHandler func(*Node, *Message.Message) error
type SyncMessageHandler func(*Node, *Message.Message) (string, error)
type CustomCommandHandler func(*Node, []string) error

type HTTPComponent interface {
	GetHTTPRequestHandlers() map[string]HTTPRequestHandler
	GetHTTPComponentConfig() Config.HTTP
}
type HTTPRequestHandler func(w http.ResponseWriter, r *http.Request)

type WebsocketComponent interface {
	GetWebsocketMessageHandlers() map[string]WebsocketMessageHandler
	OnConnectHandler(*Node, *WebsocketClient)
	OnDisconnectHandler(*Node, *WebsocketClient)
	GetWebsocketComponentConfig() Config.Websocket
}
type WebsocketMessageHandler func(*Node, *WebsocketClient, *Message.Message) error

type WebsocketApplication interface {
	Application
	WebsocketComponent
}
type HTTPApplication interface {
	Application
	HTTPComponent
}
type WebsocketHTTPApplication interface {
	Application
	WebsocketComponent
	HTTPComponent
}
