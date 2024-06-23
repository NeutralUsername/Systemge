package Client

import (
	"Systemge/Message"
	"Systemge/Resolution"
	"Systemge/Utilities"
)

func (client *Client) AddSyncTopicRemotely(brokerAddress, nameIndication, tlsCertificate, topic string) error {
	_, err := Utilities.TcpOneTimeExchange(brokerAddress, nameIndication, tlsCertificate, Message.NewAsync("addSyncTopic", client.GetName(), topic), DEFAULT_TCP_TIMEOUT)
	if err != nil {
		return Utilities.NewError("Error exchanging messages with broker", err)
	}
	return nil
}

func (client *Client) RemoveSyncTopicRemotely(brokerAddress, nameIndication, tlsCertificate, topic string) error {
	_, err := Utilities.TcpOneTimeExchange(brokerAddress, nameIndication, tlsCertificate, Message.NewAsync("removeSyncTopic", client.GetName(), topic), DEFAULT_TCP_TIMEOUT)
	if err != nil {
		return Utilities.NewError("Error exchanging messages with broker", err)
	}
	return nil
}

func (client *Client) AddAsyncTopicRemotely(brokerAddress, nameIndication, tlsCertificate, topic string) error {
	_, err := Utilities.TcpOneTimeExchange(brokerAddress, nameIndication, tlsCertificate, Message.NewAsync("addAsyncTopic", client.GetName(), topic), DEFAULT_TCP_TIMEOUT)
	if err != nil {
		return Utilities.NewError("Error exchanging messages with broker", err)
	}
	return nil
}

func (client *Client) RemoveAsyncTopicRemotely(brokerAddress, nameIndication, tlsCertificate, topic string) error {
	_, err := Utilities.TcpOneTimeExchange(brokerAddress, nameIndication, tlsCertificate, Message.NewAsync("removeAsyncTopic", client.GetName(), topic), DEFAULT_TCP_TIMEOUT)
	if err != nil {
		return Utilities.NewError("Error exchanging messages with broker", err)
	}
	return nil
}

func (client *Client) RegisterTopicRemotely(resolverAddress, nameIndication, tlsCertificate, brokerName, topic string) error {
	_, err := Utilities.TcpOneTimeExchange(resolverAddress, nameIndication, tlsCertificate, Message.NewAsync("registerTopics", client.GetName(), brokerName+" "+topic), DEFAULT_TCP_TIMEOUT)
	if err != nil {
		return Utilities.NewError("Error exchanging messages with topic resolution server", err)
	}
	return nil
}

func (client *Client) UnregisterTopicRemotely(resolverAddress, nameIndication, tlsCertificate, topic string) error {
	_, err := Utilities.TcpOneTimeExchange(resolverAddress, nameIndication, tlsCertificate, Message.NewAsync("unregisterTopics", client.GetName(), topic), DEFAULT_TCP_TIMEOUT)
	if err != nil {
		return Utilities.NewError("Error exchanging messages with topic resolution server", err)
	}
	return nil
}

func (client *Client) RegisterBrokerRemotely(resolverAddress, nameIndication, tlsCertificate string, resolution *Resolution.Resolution) error {
	_, err := Utilities.TcpOneTimeExchange(resolverAddress, nameIndication, tlsCertificate, Message.NewAsync("registerBroker", client.GetName(), resolution.Marshal()), DEFAULT_TCP_TIMEOUT)
	if err != nil {
		return Utilities.NewError("Error exchanging messages with broker", err)
	}
	return nil
}

func (client *Client) UnregisterBrokerRemotely(resolverAddress, nameIndication, tlsCertificate, brokerName string) error {
	_, err := Utilities.TcpOneTimeExchange(resolverAddress, nameIndication, tlsCertificate, Message.NewAsync("unregisterBroker", client.GetName(), brokerName), DEFAULT_TCP_TIMEOUT)
	if err != nil {
		return Utilities.NewError("Error exchanging messages with broker", err)
	}
	return nil
}
