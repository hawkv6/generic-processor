package processor

import (
	"fmt"
	"strings"
	"time"

	"github.com/hawkv6/generic-processor/pkg/config"
	"github.com/hawkv6/generic-processor/pkg/input"
	"github.com/hawkv6/generic-processor/pkg/logging"
	"github.com/hawkv6/generic-processor/pkg/message"
	"github.com/hawkv6/generic-processor/pkg/output"
	"github.com/sirupsen/logrus"
)

type TelemetryToArangoProcessor struct {
	log                      *logrus.Entry
	name                     string
	config                   config.TelemetryToArangoProcessorConfig
	inputResources           map[string]input.InputResource
	outputResources          map[string]output.OutputResource
	deactivatedIpv6Addresses map[string]string
	activeIpv6Addresses      map[string]string
	quitChan                 chan struct{}
}

func NewTelemetryToArangoProcessor(config config.TelemetryToArangoProcessorConfig, inputResources map[string]input.InputResource, outputResources map[string]output.OutputResource) *TelemetryToArangoProcessor {
	return &TelemetryToArangoProcessor{
		log:                      logging.DefaultLogger.WithField("subsystem", Subsystem),
		name:                     config.Name,
		config:                   config,
		inputResources:           inputResources,
		outputResources:          outputResources,
		deactivatedIpv6Addresses: make(map[string]string),
		activeIpv6Addresses:      make(map[string]string),
		quitChan:                 make(chan struct{}),
	}
}

func (processor *TelemetryToArangoProcessor) Init() error {
	processor.log.Infof("Initializing TelemetryToArangoProcessor '%s'", processor.name)
	return nil
}

func (processor *TelemetryToArangoProcessor) processIpv6Message(msg *message.IPv6Message) {
	processor.log.Debugf("Processing IPv6 message: %v", msg)
	if strings.Contains(msg.Tags.InterfaceName, "Ethernet") {
		if !msg.Fields.Delete {
			processor.activeIpv6Addresses[msg.Tags.Source+msg.Tags.InterfaceName] = msg.Fields.IPv6
		} else {
			delete(processor.activeIpv6Addresses, msg.Tags.Source+msg.Tags.InterfaceName)
		}
	}
}

func (processor *TelemetryToArangoProcessor) processInterfaceStatusMessage(msg *message.InterfaceStatusMessage) {
	processor.log.Debugf("Processing InterfaceStatus message: %v", msg)
	if strings.Contains(msg.Tags.InterfaceName, "Ethernet") {
		if msg.Fields.AdminStatus == "UP" {
			processor.activeIpv6Addresses[msg.Tags.Source+msg.Tags.InterfaceName] = processor.deactivatedIpv6Addresses[msg.Tags.Source+msg.Tags.InterfaceName]
			delete(processor.deactivatedIpv6Addresses, msg.Tags.Source+msg.Tags.InterfaceName)
		} else {
			processor.deactivatedIpv6Addresses[msg.Tags.Source+msg.Tags.InterfaceName] = processor.activeIpv6Addresses[msg.Tags.Source+msg.Tags.InterfaceName]
			delete(processor.activeIpv6Addresses, msg.Tags.Source+msg.Tags.InterfaceName)
		}
	}
}

func (processor *TelemetryToArangoProcessor) startKafkaProcessing(name string, input *input.KafkaInput, commandChan chan message.Command, resultChan chan message.Result) {
	commandChan <- message.KafkaListeningCommand{}
	processor.log.Debugf("Starting Processing '%s' input messages", name)
	for {
		select {
		case msg := <-resultChan:
			processor.log.Debugf("Received message from '%s' input: %v", name, msg)
			switch msgType := msg.(type) {
			case *message.IPv6Message:
				processor.processIpv6Message(msgType)
			case *message.InterfaceStatusMessage:
				processor.processInterfaceStatusMessage(msgType)
			}
		case <-processor.quitChan:
			processor.log.Infof("Stopping Processing '%s' input messages", name)
			return
		}
	}
}

func (processor *TelemetryToArangoProcessor) sendCommands(name string, commandChan chan message.Command) {
	for _, mode := range processor.config.Modes {
		for inputName, inputOption := range mode.InputOptions {
			if inputName == name {
				commandChan <- message.InfluxQueryCommand{
					Measurement:   inputOption.Measurement,
					Field:         inputOption.Field,
					Method:        inputOption.Method,
					GroupBy:       inputOption.GroupBy,
					Interval:      processor.config.Interval,
					OutputOptions: mode.OutputOptions,
				}
			}
		}
	}
}

func (processor *TelemetryToArangoProcessor) startSchedulingInfluxCommands(name string, input *input.InfluxInput, commandChan chan message.Command) {
	processor.log.Infof("Starting scheduling Influx commands for '%s' input every %d seconds", name, processor.config.Interval)
	ticker := time.NewTicker(time.Duration(processor.config.Interval) * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		processor.sendCommands(name, commandChan)
	}
}

func (processor *TelemetryToArangoProcessor) getLocalLinkIp(tags map[string]string) (string, error) {
	sourceTag, ok := tags["source"]
	if !ok {
		err := fmt.Errorf("Received unknown source tag: %v", sourceTag)
		processor.log.Errorln(err)
		return "", err
	}
	interfaceNameTag, ok := tags["interface_name"]
	if !ok {
		err := fmt.Errorf("Received unknown interface_name tag: %v", interfaceNameTag)
		processor.log.Errorln(err)
		return "", err
	}
	ipv6Address := processor.activeIpv6Addresses[sourceTag+interfaceNameTag]
	if ipv6Address == "" {
		err := fmt.Errorf("Received empty IPv6 address for source: %v and interface_name: %v", sourceTag, interfaceNameTag)
		processor.log.Errorln(err)
		return "", err
	}
	return ipv6Address, nil
}

func (processor *TelemetryToArangoProcessor) sendUpdateCommands(outputOption config.OutputOption, tags map[string]string, commandChan chan message.Command, value interface{}) {
	filterBy := make(map[string]interface{})
	for index := range outputOption.FilterBy {
		if filterKey := outputOption.FilterBy[index]; filterKey == "local_link_ip" {
			localLinkIp, err := processor.getLocalLinkIp(tags)
			if err != nil {
				processor.log.Errorln(err)
				return
			}
			filterBy[filterKey] = localLinkIp
		} else {
			processor.log.Errorf("Received not supported filter_by value: %v", filterKey)
		}
		commandChan <- message.ArangoUpdateCommand{
			Collection: outputOption.Collection,
			FilterBy:   filterBy,
			Field:      outputOption.Field,
			Value:      value,
			Index:      outputOption.Index,
		}
	}
}

func (processor *TelemetryToArangoProcessor) processInfluxResultMessage(name string, msg message.Result) {
	processor.log.Debugf("Received message from '%s' input: %v", name, msg)
	switch msgType := msg.(type) {
	case message.InfluxResultMessage:
		for outputName, outputResource := range processor.outputResources {
			outputOption, ok := msgType.OutputOptions[outputName]
			if !ok {
				processor.log.Errorf("Received unknown output name: %v", outputName)
				continue
			}
			if outputOption.Method == "update" {
				processor.sendUpdateCommands(outputOption, msgType.Tags, outputResource.CommandChan, msgType.Value)
			} else {
				processor.log.Errorf("Received unknown output method: %v", msgType.OutputOptions[outputName].Method)
			}
		}
	default:
		processor.log.Errorf("Received unknown message type: %v", msgType)
	}
}

func (processor *TelemetryToArangoProcessor) startInfluxProcessing(name string, input *input.InfluxInput, commandChan chan message.Command, resultChan chan message.Result) {
	processor.log.Debugf("Starting Processing '%s' input messages", name)
	go processor.startSchedulingInfluxCommands(name, input, commandChan)

	for {
		select {
		case msg := <-resultChan:
			processor.processInfluxResultMessage(name, msg)
		case <-processor.quitChan:
			processor.log.Infof("Stopping Processing '%s' input messages", name)
			return
		}
	}
}

func (processor *TelemetryToArangoProcessor) Start() {
	processor.log.Infof("Starting TelemetryToArangoProcessor '%s'", processor.name)
	for name, inputResource := range processor.inputResources {
		switch input := inputResource.Input.(type) {
		case *input.KafkaInput:
			go processor.startKafkaProcessing(name, input, inputResource.CommandChan, inputResource.ResultChan)
		case *input.InfluxInput:
			go processor.startInfluxProcessing(name, input, inputResource.CommandChan, inputResource.ResultChan)
		}
	}
}

func (processor *TelemetryToArangoProcessor) Stop() {
	close(processor.quitChan)
	processor.log.Infof("Stopping TelemetryToArangoProcessor '%s'", processor.name)
}
