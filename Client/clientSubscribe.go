package Client

import (
	"Systemge/Message"
	"Systemge/Utilities"
)

// subscribes to a topic from the provided server connection
func (client *Client) subscribeTopic(brokerConnection *brokerConnection, topic string) error {
	message := Message.NewSync("subscribe", client.config.Name, topic)
	responseChannel, err := client.addMessageWaitingForResponse(message)
	if err != nil {
		return Utilities.NewError("Error adding message to waiting for response map", err)
	}
	err = brokerConnection.send(message)
	if err != nil {
		client.removeMessageWaitingForResponse(message.GetSyncRequestToken(), responseChannel)
		return Utilities.NewError("Failed to send message with topic \""+message.GetTopic()+"\" to message broker server", err)
	}
	response, err := client.receiveSyncResponse(message, responseChannel)
	if err != nil {
		return Utilities.NewError("Error sending subscription sync request", err)
	}
	if response.GetTopic() != "subscribed" {
		return Utilities.NewError("Invalid response topic \""+response.GetTopic()+"\"", nil)
	}
	return nil
}