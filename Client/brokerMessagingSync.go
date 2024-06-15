package Client

import (
	"Systemge/Error"
	"Systemge/Message"
)

// resolves the broker address for the provided topic and sends the sync message to the broker responsible for the topic and waits for a response.
func (client *Client) SyncMessage(topic, origin, payload string) (*Message.Message, error) {
	message := Message.NewSync(topic, origin, payload)
	if message.GetSyncRequestToken() == "" {
		return nil, Error.New("SyncRequestToken not set", nil)
	}
	brokerConnection, err := client.getBrokerConnectionForTopic(message.GetTopic())
	if err != nil {
		return nil, Error.New("Error resolving broker address for topic \""+message.GetTopic()+"\"", err)
	}
	err = brokerConnection.send(message)
	if err != nil {
		return nil, Error.New("Error sending sync request message", err)
	}
	return client.receiveSyncResponse(message)
}

func (client *Client) receiveSyncResponse(message *Message.Message) (*Message.Message, error) {
	responseChannel := client.addMessageWaitingForResponse(message)
	select {
	case response := <-responseChannel:
		client.removeMessageWaitingForResponse(message.GetSyncRequestToken())
		if response.GetTopic() == "error" {
			return nil, Error.New(response.GetPayload(), nil)
		}
		return response, nil
	case <-client.stopChannel:
		return nil, Error.New("client stopped", nil)
	}
}
func (client *Client) addMessageWaitingForResponse(message *Message.Message) chan *Message.Message {
	client.mapOperationMutex.Lock()
	defer client.mapOperationMutex.Unlock()
	responseChannel := make(chan *Message.Message)
	client.messagesWaitingForResponse[message.GetSyncRequestToken()] = responseChannel
	return responseChannel
}
func (client *Client) removeMessageWaitingForResponse(syncRequestToken string) {
	client.mapOperationMutex.Lock()
	defer client.mapOperationMutex.Unlock()
	delete(client.messagesWaitingForResponse, syncRequestToken)
}