package processor

import (
	"fmt"
	"strings"

	"github.com/hawkv6/generic-processor/pkg/logging"
	"github.com/hawkv6/generic-processor/pkg/message"
	"github.com/sirupsen/logrus"
)

type KafkaProcessor interface {
	Start(string, chan message.Command, chan message.Result)
	GetLocalLinkIp(map[string]string) (string, error)
	Stop()
}

type KafkaOpenConfigProcessor struct {
	log                      *logrus.Entry
	deactivatedIpv6Addresses map[string]string
	activeIpv6Addresses      map[string]string
	quitChan                 chan struct{}
}

func NewKafkaOpenConfigProcessor() *KafkaOpenConfigProcessor {
	return &KafkaOpenConfigProcessor{
		log:                      logging.DefaultLogger.WithField("subsystem", Subsystem),
		deactivatedIpv6Addresses: make(map[string]string),
		activeIpv6Addresses:      make(map[string]string),
		quitChan:                 make(chan struct{}),
	}
}

func (processor *KafkaOpenConfigProcessor) processIpv6Message(msg *message.IPv6Message) {
	processor.log.Debugf("Processing IPv6 message for IPv6 %s", msg.Fields.IPv6)
	if strings.Contains(msg.Tags.InterfaceName, "Ethernet") {
		if !msg.Fields.Delete {
			processor.activeIpv6Addresses[msg.Tags.Source+msg.Tags.InterfaceName] = msg.Fields.IPv6
		} else {
			delete(processor.activeIpv6Addresses, msg.Tags.Source+msg.Tags.InterfaceName)
		}
	}
}

func (processor *KafkaOpenConfigProcessor) processInterfaceStatusMessage(msg *message.InterfaceStatusMessage) {
	processor.log.Debugf("Processing Interface Status message: %s %s", msg.Tags.Source, msg.Tags.InterfaceName)
	if strings.Contains(msg.Tags.InterfaceName, "Ethernet") {
		if msg.Fields.AdminStatus == "UP" {
			processor.log.Debugf("Interface '%s' from router '%s' changed to UP", msg.Tags.InterfaceName, msg.Tags.Source)
			if _, ok := processor.activeIpv6Addresses[msg.Tags.Source+msg.Tags.InterfaceName]; !ok {
				processor.activeIpv6Addresses[msg.Tags.Source+msg.Tags.InterfaceName] = processor.deactivatedIpv6Addresses[msg.Tags.Source+msg.Tags.InterfaceName]
				delete(processor.deactivatedIpv6Addresses, msg.Tags.Source+msg.Tags.InterfaceName)
			}
		} else {
			processor.log.Debugf("Interface '%s' from router '%s' changed to DOWN", msg.Tags.InterfaceName, msg.Tags.Source)
			if _, ok := processor.deactivatedIpv6Addresses[msg.Tags.Source+msg.Tags.InterfaceName]; !ok {
				processor.deactivatedIpv6Addresses[msg.Tags.Source+msg.Tags.InterfaceName] = processor.activeIpv6Addresses[msg.Tags.Source+msg.Tags.InterfaceName]
				delete(processor.activeIpv6Addresses, msg.Tags.Source+msg.Tags.InterfaceName)
			}
		}
	}
}

func (processor *KafkaOpenConfigProcessor) Start(name string, commandChan chan message.Command, resultChan chan message.Result) {
	commandChan <- message.KafkaListeningCommand{}
	processor.log.Debugf("Starting processing '%s' input messages", name)
	for {
		select {
		case msg := <-resultChan:
			processor.log.Debugf("Received message from '%s' input", name)
			switch msgType := msg.(type) {
			case *message.IPv6Message:
				processor.processIpv6Message(msgType)
			case *message.InterfaceStatusMessage:
				processor.processInterfaceStatusMessage(msgType)
			default:
				processor.log.Errorln("Received unknown message type: ", msgType)
			}
		case <-processor.quitChan:
			processor.log.Infof("Stopping processing '%s' input messages", name)
			return
		}
	}
}

func (processor *KafkaOpenConfigProcessor) GetLocalLinkIp(tags map[string]string) (string, error) {
	sourceTag, ok := tags["source"]
	if !ok {
		return "", fmt.Errorf("Received unknown source tag: %v", sourceTag)
	}
	ipv6Address := ""
	for key, value := range tags {
		if key != "source" {
			ipv6Address := processor.activeIpv6Addresses[sourceTag+value]
			if ipv6Address != "" {
				return ipv6Address, nil
			}
		}
	}
	if ipv6Address == "" {
		return "", fmt.Errorf("Received empty IPv6 address for source: %v tried combination with all received tags: %v", sourceTag, tags)
	}
	return ipv6Address, nil
}

func (processor *KafkaOpenConfigProcessor) Stop() {
	processor.log.Info("Stopping Kafka OpenConfig Processor")
	close(processor.quitChan)
}
