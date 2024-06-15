package Module

import (
	"Systemge/Broker"
	"Systemge/Utilities"
	"strings"
)

func NewBroker(name, tlsListenerPort, tlsCertPath, tlsKeyPath, loggerPath string, asyncTopics []string, syncTopics []string) *Broker.Server {
	logger := Utilities.NewLogger(loggerPath)
	broker := Broker.New(name, tlsListenerPort, tlsCertPath, tlsKeyPath, logger)
	for _, topic := range asyncTopics {
		broker.AddAsyncTopics(topic)
	}
	for _, topic := range syncTopics {
		broker.AddSyncTopics(topic)
	}
	return broker
}

func NewBrokerFromConfig(sytemgeConfigPath string, errorLogPath string) *Broker.Server {
	if !Utilities.FileExists(sytemgeConfigPath) {
		panic("provided file does not exist")
	}
	filename := Utilities.GetFileName(sytemgeConfigPath)
	fileNameSegments := strings.Split(Utilities.GetFileName(sytemgeConfigPath), ".")
	name := ""
	if len(fileNameSegments) != 2 {
		name = filename
	} else {
		name = fileNameSegments[0]
	}
	tlsListenerPort := ""
	tlsCertPath := ""
	tlsKeyPath := ""
	asyncTopics := []string{}
	syncTopics := []string{}
	for i, line := range Utilities.SplitLines(Utilities.GetFileContent(sytemgeConfigPath)) {
		if len(line) == 0 {
			continue
		}
		switch i {
		case 0:
			if len(line) < 2 {
				panic("error reading file. missing config type")
			}
			segments := strings.Split(line, " ")
			if len(segments) != 2 {
				panic("error reading file. incomplete config type")
			}
			if segments[0] != "#" {
				panic("error reading file. missing config type identifier (#)")
			}
			if segments[1] != "broker" {
				panic("error reading file. invalid config type (want: broker)")
			}
			continue
		case 1:
			if len(line) < 2 {
				panic("error reading file. Missing port number")
			}
			if line[0] != ':' {
				panic("error reading file. Missing port number")
			}
			tlsListenerPort = line
		case 2:
			lineSegments := strings.Split(line, " ")
			if len(lineSegments) != 2 {
				panic("error reading file. invalid cert or key line")
			}
			if lineSegments[0] == "cert" {
				tlsCertPath = lineSegments[1]
			} else if lineSegments[0] == "key" {
				tlsKeyPath = lineSegments[1]
			} else {
				panic("error reading file. invalid cert or key type")
			}
		case 3:
			lineSegments := strings.Split(line, " ")
			if len(lineSegments) != 2 {
				panic("error reading file. invalid cert or key line")
			}
			if lineSegments[0] == "cert" {
				tlsCertPath = lineSegments[1]
			} else if lineSegments[0] == "key" {
				tlsKeyPath = lineSegments[1]
			} else {
				panic("error reading file. invalid cert or key type")
			}
		default:
			lineSegments := strings.Split(line, " ")
			if len(lineSegments) != 2 {
				panic("error reading file. invalid topic line")
			}
			if lineSegments[1] == "sync" {
				syncTopics = append(syncTopics, lineSegments[0])
			} else if lineSegments[1] == "async" {
				asyncTopics = append(asyncTopics, lineSegments[0])
			} else {
				panic("error reading file. invalid topic type")
			}
		}
	}
	return NewBroker(name, tlsListenerPort, tlsCertPath, tlsKeyPath, errorLogPath, asyncTopics, syncTopics)
}