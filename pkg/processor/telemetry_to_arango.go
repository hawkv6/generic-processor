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
	log             *logrus.Entry
	name            string
	config          config.TelemetryToArangoProcessorConfig
	inputResources  map[string]input.InputResource
	outputResources map[string]output.OutputResource
	quitChan        chan struct{}
	kafkaProcessor  KafkaProcessor
	normalizer      Normalizer
}

func NewTelemetryToArangoProcessor(config config.TelemetryToArangoProcessorConfig, inputResources map[string]input.InputResource, outputResources map[string]output.OutputResource, kafkaOpenConfigProcessor KafkaProcessor, normalizer Normalizer) *TelemetryToArangoProcessor {
	return &TelemetryToArangoProcessor{
		log:             logging.DefaultLogger.WithField("subsystem", Subsystem),
		name:            config.Name,
		config:          config,
		inputResources:  inputResources,
		outputResources: outputResources,
		quitChan:        make(chan struct{}),
		kafkaProcessor:  kafkaOpenConfigProcessor,
		normalizer:      normalizer,
	}
}

func (processor *TelemetryToArangoProcessor) sendInfluxCommands(name string, commandChan chan message.Command) {
	for _, mode := range processor.config.Modes {
		for inputName, inputOption := range mode.InputOptions {
			if inputName == name {
				commandChan <- message.InfluxQueryCommand{
					Measurement:    inputOption.Measurement,
					Field:          inputOption.Field,
					Transformation: inputOption.Transformation,
					Method:         inputOption.Method,
					GroupBy:        inputOption.GroupBy,
					Interval:       processor.config.Interval,
					OutputOptions:  mode.OutputOptions,
				}
			}
		}
	}
}

func (processor *TelemetryToArangoProcessor) startSchedulingInfluxCommands(name string, commandChan chan message.Command) {
	processor.log.Infof("Starting scheduling Influx commands for '%s' input every %d seconds", name, processor.config.Interval)
	ticker := time.NewTicker(time.Duration(processor.config.Interval) * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			processor.log.Infoln("Sending Influx commands")
			processor.sendInfluxCommands(name, commandChan)
		case <-processor.quitChan:
			processor.log.Infof("Stopping scheduling Influx commands for '%s' input", name)
			return
		}
	}
}

func (processor *TelemetryToArangoProcessor) sendArangoUpdateCommands(arangoUpdateCommands map[string]map[string]message.ArangoUpdateCommand) {
	for name, outputResource := range processor.outputResources {
		if arangoUpdateCommand, ok := arangoUpdateCommands[name]; ok {
			for collection, command := range arangoUpdateCommand {
				outputResource.CommandChan <- command
				processor.log.Debugf("Sending Arango update command for collection: %s", collection)
			}
		}
	}
}

func (processor *TelemetryToArangoProcessor) checkValidInterfaceName(tags map[string]string) bool {
	for _, value := range tags {
		if strings.Contains(value, "Ethernet") {
			return true
		}
	}
	return false
}

// TODO refactor this function
func (processor *TelemetryToArangoProcessor) createArangoUpdateCommands(outputOption config.OutputOption, results []message.InfluxResult, command message.ArangoUpdateCommand) {
	for _, result := range results {
		if !processor.checkValidInterfaceName(result.Tags) {
			processor.log.Debugf("Received not supported interface name: %v", result.Tags)
			continue
		}
		filterBy := make(map[string]interface{})
		var key string
		for index := range outputOption.FilterBy {
			if filterKey := outputOption.FilterBy[index]; filterKey == "local_link_ip" {
				localLinkIp, err := processor.kafkaProcessor.GetLocalLinkIp(result.Tags)
				if err != nil {
					processor.log.Errorln(err)
					continue
				}
				filterBy[filterKey] = localLinkIp
				key = localLinkIp
			} else {
				processor.log.Errorf("Received not supported filter_by value: %v", filterKey)
			}
		}
		if key == "" {
			processor.log.Errorf("Received empty key for filter_by: %v", outputOption.FilterBy)
			continue
		}
		var update message.ArangoUpdate
		if _, ok := command.Updates[key]; !ok {
			update = message.ArangoUpdate{
				Tags:   result.Tags,
				Fields: make([]string, 0),
				Values: make([]float64, 0),
			}
		} else {
			update = command.Updates[key]
		}
		update.Fields = append(update.Fields, outputOption.Field)
		update.Values = append(update.Values, result.Value)
		if _, ok := processor.config.Normalization.FieldMappings[outputOption.Field]; ok {
			processor.normalizer.AddDataToNormalize(outputOption.Field, result.Value)
		}
		command.Updates[key] = update
	}
	if normalizationField, ok := processor.config.Normalization.FieldMappings[outputOption.Field]; ok {
		statisticalData, err := processor.normalizer.NormalizeValues(command.Updates, outputOption.Field, normalizationField)
		if err != nil {
			processor.log.Errorln(err)
		} else {
			command.StatisticalData[fmt.Sprintf("%s_%s", outputOption.Field, "q1")] = statisticalData.q1
			command.StatisticalData[fmt.Sprintf("%s_%s", outputOption.Field, "q3")] = statisticalData.q3
			command.StatisticalData[fmt.Sprintf("%s_%s", outputOption.Field, "iqr")] = statisticalData.interQuartileRange
			command.StatisticalData[fmt.Sprintf("%s_%s", outputOption.Field, "lower_fence")] = statisticalData.lowerFence
			command.StatisticalData[fmt.Sprintf("%s_%s", outputOption.Field, "upper_fence")] = statisticalData.upperFence
		}
	}
}

// todo refactor this method
func (processor *TelemetryToArangoProcessor) processInfluxResultMessage(name string, messages []message.Result) {
	arangoUpdateCommands := make(map[string]map[string]message.ArangoUpdateCommand)

	processor.log.Debugf("Received %d different types of result messages from '%s' input", len(messages), name)
	for _, msg := range messages {
		switch msgType := msg.(type) {
		case message.InfluxResultMessage:
			for name, outputOption := range msgType.OutputOptions {
				switch output := processor.outputResources[name].Output.(type) {
				case *output.ArangoOutput:
					if outputOption.Method == "update" {
						if _, ok := arangoUpdateCommands[name]; !ok {
							arangoUpdateCommands[name] = make(map[string]message.ArangoUpdateCommand)
						}
						if _, ok := arangoUpdateCommands[name][outputOption.Collection]; !ok {
							arangoUpdateCommands[name][outputOption.Collection] = *message.NewArangoUpdateCommand(outputOption.Collection)
						}
						processor.normalizer.ResetNormalizationData()
						processor.createArangoUpdateCommands(outputOption, msgType.Results, arangoUpdateCommands[name][outputOption.Collection])
					}
				default:
					processor.log.Errorf("Received not yet supported output type: %v", output)
				}

			}
		default:
			processor.log.Errorf("Received unknown message type from influx: %v", msgType)
		}
	}
	processor.sendArangoUpdateCommands(arangoUpdateCommands)
}

func (processor *TelemetryToArangoProcessor) startInfluxProcessing(name string, commandChan chan message.Command, resultChan chan message.Result) {
	processor.log.Debugf("Starting Processing '%s' input messages", name)
	go processor.startSchedulingInfluxCommands(name, commandChan)
	modeCount := len(processor.config.Modes)
	count := 0
	messages := make([]message.Result, modeCount)
	for {
		select {
		case msg := <-resultChan:
			if count < modeCount-1 {
				messages[count] = msg
				count++
			} else {
				messages[count] = msg
				processor.processInfluxResultMessage(name, messages)
				count = 0
			}
		case <-processor.quitChan:
			processor.log.Infof("Stopping Processing '%s' input messages", name)
			return
		}
	}
}

// todo refactor this method
func (processor *TelemetryToArangoProcessor) processArangoResultMessages(kafkaOutputs map[string]output.OutputResource, arangoResultChannels map[string]chan message.Result) {
	for {
		select {
		case <-processor.quitChan:
			return
		default:
			for _, arangoResultChannel := range arangoResultChannels {
				for msg := range arangoResultChannel {
					switch msgType := msg.(type) {
					case message.ArangoEventNotificationMessage:
						processor.log.Debugf("Received %d Arango event messages", len(msgType.EventMessages))
						commandMessage := message.KafkaEventCommand{}
						commandMessage.Updates = make([]message.KafkaEventMessage, len(msgType.EventMessages))
						for index, event := range msgType.EventMessages {
							commandMessage.Updates[index] = message.KafkaEventMessage{
								TopicType: event.TopicType,
								Key:       event.Key,
								Id:        event.Id,
								Action:    "update",
							}
						}
						for _, output := range kafkaOutputs {
							output.CommandChan <- commandMessage
						}
					case message.ArangoNormalizationMessage:
						processor.log.Debugf("Received %d Arango normalization messages", len(msgType.NormalizationMessages))
						commandMessage := message.KafkaNormalizationCommand{}
						commandMessage.Updates = make([]message.KafkaNormalizationMessage, len(msgType.NormalizationMessages))
						for index, normalization := range msgType.NormalizationMessages {
							commandMessage.Updates[index] = message.KafkaNormalizationMessage{
								Measurement: msgType.Measurement,
								Tags:        normalization.Tags,
								Fields:      normalization.Fields,
							}
						}
						for _, output := range kafkaOutputs {
							output.CommandChan <- commandMessage
						}
					default:
						processor.log.Errorf("Received unknown message type: %v", msgType)
					}
				}
			}
		}
	}
}

func (processor *TelemetryToArangoProcessor) Start() {
	processor.log.Infof("Starting TelemetryToArangoProcessor '%s'", processor.name)
	for name, inputResource := range processor.inputResources {
		switch inputResource.Input.(type) {
		case *input.KafkaInput:
			go processor.kafkaProcessor.Start(name, inputResource.CommandChan, inputResource.ResultChan)
		case *input.InfluxInput:
			go processor.startInfluxProcessing(name, inputResource.CommandChan, inputResource.ResultChan)
		}
	}

	arangoResultChannels := make(map[string]chan message.Result)
	kafkaOutputs := make(map[string]output.OutputResource)
	for name, outputResource := range processor.outputResources {
		switch outputResource.Output.(type) {
		case *output.ArangoOutput:
			arangoResultChannels[name] = outputResource.ResultChan
		case *output.KafkaOutput:
			kafkaOutputs[name] = outputResource
		}
	}
	go processor.processArangoResultMessages(kafkaOutputs, arangoResultChannels)
}

func (processor *TelemetryToArangoProcessor) Stop() {
	processor.log.Infof("Stopping TelemetryToArangoProcessor '%s'", processor.name)
	processor.kafkaProcessor.Stop()
	close(processor.quitChan)
}
