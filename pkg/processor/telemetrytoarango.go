package processor

import (
	"time"

	"github.com/hawkv6/generic-processor/pkg/config"
	"github.com/hawkv6/generic-processor/pkg/input"
	"github.com/hawkv6/generic-processor/pkg/logging"
	"github.com/hawkv6/generic-processor/pkg/message"
	"github.com/hawkv6/generic-processor/pkg/output"
	"github.com/sirupsen/logrus"
)

type TelemetryToArangoProcessor struct {
	log                 *logrus.Entry
	name                string
	config              config.TelemetryToArangoProcessorConfig
	inputResources      map[string]input.InputResource
	outputResources     map[string]output.OutputResource
	ipv6Addresses       map[string]string
	activeIpv6Addresses map[string]string
	quitChan            chan struct{}
}

func NewTelemetryToArangoProcessor(config config.TelemetryToArangoProcessorConfig, inputResources map[string]input.InputResource, outputResources map[string]output.OutputResource) *TelemetryToArangoProcessor {
	return &TelemetryToArangoProcessor{
		log:                 logging.DefaultLogger.WithField("subsystem", Subsystem),
		name:                config.Name,
		config:              config,
		inputResources:      inputResources,
		outputResources:     outputResources,
		ipv6Addresses:       make(map[string]string),
		activeIpv6Addresses: make(map[string]string),
		quitChan:            make(chan struct{}),
	}
}

func (processor *TelemetryToArangoProcessor) Init() error {
	processor.log.Infof("Initializing TelemetryToArangoProcessor '%s'", processor.name)
	return nil
}

func (processor *TelemetryToArangoProcessor) processIpv6Message(msg *message.IPv6Message) {
	processor.log.Debugf("Processing IPv6 message: %v", msg)
	if !msg.Fields.Delete {
		processor.ipv6Addresses[msg.Tags.Source+msg.Tags.InterfaceName] = msg.Fields.IPv6
	} else {
		delete(processor.ipv6Addresses, msg.Tags.Source+msg.Tags.InterfaceName)
		delete(processor.activeIpv6Addresses, msg.Tags.Source+msg.Tags.InterfaceName)
	}
}

func (processor *TelemetryToArangoProcessor) processInterfaceStatusMessage(msg *message.InterfaceStatusMessage) {
	processor.log.Debugf("Processing InterfaceStatus message: %v", msg)
	if msg.Fields.AdminStatus == "UP" {
		processor.activeIpv6Addresses[msg.Tags.Source+msg.Tags.InterfaceName] = processor.ipv6Addresses[msg.Tags.Source+msg.Tags.InterfaceName]
	} else {
		delete(processor.activeIpv6Addresses, msg.Tags.Source+msg.Tags.InterfaceName)
		// here the ls link is deleted from the database
	}
}

func (processor *TelemetryToArangoProcessor) startKafkaProcessing(name string, input *input.KafkaInput, commandChan chan message.Command, resultChan chan message.ResultMessage) {
	commandChan <- message.StartListeningCommand{}
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
				commandChan <- message.InfluxCommand{
					Measurement: inputOption.Measurement,
					Field:       inputOption.Field,
					Method:      inputOption.Method,
					GroupBy:     inputOption.GroupBy,
					Interval:    processor.config.Interval,
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

func (processor *TelemetryToArangoProcessor) processResultMessage(name string, msg message.ResultMessage) {
	processor.log.Debugf("Received message from '%s' input: %v", name, msg)
	switch msgType := msg.(type) {
	case *message.InfluxResultMessage:
		processor.log.Debugf("Received InfluxResultMessage: %v", msgType)
		// here the ls link is created in the database
	}
}

func (processor *TelemetryToArangoProcessor) startInfluxProcessing(name string, input *input.InfluxInput, commandChan chan message.Command, resultChan chan message.ResultMessage) {
	processor.log.Debugf("Starting Processing '%s' input messages", name)
	go processor.startSchedulingInfluxCommands(name, input, commandChan)

	for {
		select {
		case msg := <-resultChan:
			processor.processResultMessage(name, msg)
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
			go processor.startKafkaProcessing(name, input, inputResource.CommandChannel, inputResource.ResultChannel)
		case *input.InfluxInput:
			go processor.startInfluxProcessing(name, input, inputResource.CommandChannel, inputResource.ResultChannel)
		}
	}
}

func (processor *TelemetryToArangoProcessor) Stop() {
	close(processor.quitChan)
	processor.log.Infof("Stopping TelemetryToArangoProcessor '%s'", processor.name)
}
