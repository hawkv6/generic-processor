package processor

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/hawkv6/generic-processor/pkg/config"
	"github.com/hawkv6/generic-processor/pkg/input"
	"github.com/hawkv6/generic-processor/pkg/logging"
	"github.com/hawkv6/generic-processor/pkg/message"
	"github.com/hawkv6/generic-processor/pkg/output"
	"github.com/montanaflynn/stats"
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
	normalizationData        map[string][]float64
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
	processor.log.Debugf("Processing IPv6 message for IPv6 %s", msg.Fields.IPv6)
	if strings.Contains(msg.Tags.InterfaceName, "Ethernet") {
		if !msg.Fields.Delete {
			processor.activeIpv6Addresses[msg.Tags.Source+msg.Tags.InterfaceName] = msg.Fields.IPv6
		} else {
			delete(processor.activeIpv6Addresses, msg.Tags.Source+msg.Tags.InterfaceName)
		}
	}
}

func (processor *TelemetryToArangoProcessor) processInterfaceStatusMessage(msg *message.InterfaceStatusMessage) {
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

func (processor *TelemetryToArangoProcessor) startKafkaProcessing(name string, commandChan chan message.Command, resultChan chan message.Result) {
	commandChan <- message.KafkaListeningCommand{}
	processor.log.Debugf("Starting Processing '%s' input messages", name)
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
			processor.log.Infof("Stopping Processing '%s' input messages", name)
			return
		}
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

func (processor *TelemetryToArangoProcessor) getLocalLinkIp(tags map[string]string) (string, error) {
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

func (processor *TelemetryToArangoProcessor) addDataToNormalize(fieldName string, value float64) {
	if _, ok := processor.normalizationData[fieldName]; !ok {
		processor.normalizationData[fieldName] = make([]float64, 0)
	} else {
		processor.normalizationData[fieldName] = append(processor.normalizationData[fieldName], value)
	}
}

type StatisticalData struct {
	q1                 float64
	q3                 float64
	interQuartileRange float64
	lowerFence         float64
	upperFence         float64
}

func NewStatisticalData(q1, q3, iqr, lowerFence, upperFence float64) *StatisticalData {
	return &StatisticalData{
		q1:                 q1,
		q3:                 q3,
		interQuartileRange: iqr,
		lowerFence:         lowerFence,
		upperFence:         upperFence,
	}
}

func (processor *TelemetryToArangoProcessor) getFences(data stats.Float64Data, upperFence, lowerFence *float64) (*StatisticalData, error) {
	quartiles, err := stats.Quartile(data)
	if err != nil {
		return nil, fmt.Errorf("Error calculating quartiles: %v", err)
	}
	interQuartileRange := quartiles.Q3 - quartiles.Q1
	processor.log.Debugln("Q1: ", quartiles.Q1)
	processor.log.Debugln("Q2 / Median: ", quartiles.Q2)
	processor.log.Debugln("Q3: ", quartiles.Q3)
	processor.log.Debugln("Interquartile range: ", interQuartileRange)
	outliers, err := stats.QuartileOutliers(data)
	if err != nil {
		return nil, fmt.Errorf("Error calculating outliers: %v", err)
	}
	processor.log.Debugf("Outliers: %+v", outliers)

	min, err := stats.Min(data)
	if err != nil {
		return nil, fmt.Errorf("Error calculating min: %v", err)
	}

	max, err := stats.Max(data)
	if err != nil {
		return nil, fmt.Errorf("Error calculating max: %v", err)
	}

	*upperFence = math.Min(quartiles.Q3+1.5*interQuartileRange, max)
	processor.log.Debugln("Upper fence: ", *upperFence)
	*lowerFence = math.Max(quartiles.Q1-1.5*interQuartileRange, min)
	processor.log.Debugln("Lower fence: ", *lowerFence)
	return NewStatisticalData(quartiles.Q1, quartiles.Q3, interQuartileRange, *lowerFence, *upperFence), nil
}

func (processor *TelemetryToArangoProcessor) getNormalizedValue(value float64, upperFence, lowerFence float64) float64 {
	normalizedValue := value
	if upperFence != lowerFence {
		normalizedValue = (value - lowerFence) / (upperFence - lowerFence)
		if normalizedValue <= 0 {
			normalizedValue = 0.0000000001
		} else if normalizedValue > 1 {
			normalizedValue = 1
		}
	} else {
		processor.log.Warnf("Upper fence and lower fence are equal: %v", upperFence)
	}
	return normalizedValue
}

func (processor *TelemetryToArangoProcessor) normalizeValues(updates map[string]message.ArangoUpdate, inputField, normalizationField string) (*StatisticalData, error) {
	if _, ok := processor.normalizationData[inputField]; ok {
		data := stats.LoadRawData(processor.normalizationData[inputField])
		upperFence, lowerFence := 0.0, 0.0
		statisticalData, err := processor.getFences(data, &upperFence, &lowerFence)
		if err != nil {
			return nil, err
		}
		for key, update := range updates {
			normalizedValue := processor.getNormalizedValue(update.Values[len(update.Values)-1], upperFence, lowerFence)
			update.Fields = append(update.Fields, normalizationField)
			update.Values = append(update.Values, normalizedValue)
			updates[key] = update
		}
		return statisticalData, nil
	} else {
		return nil, fmt.Errorf("No data to normalize for field: %v", inputField)
	}

}

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
				localLinkIp, err := processor.getLocalLinkIp(result.Tags)
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
			processor.addDataToNormalize(outputOption.Field, result.Value)
		}
		command.Updates[key] = update
	}
	if normalizationField, ok := processor.config.Normalization.FieldMappings[outputOption.Field]; ok {
		// processor.normalizeValues(command.Updates, outputOption.Field, normalizationField)
		statisticalData, err := processor.normalizeValues(command.Updates, outputOption.Field, normalizationField)
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
						processor.normalizationData = make(map[string][]float64)
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
			go processor.startKafkaProcessing(name, inputResource.CommandChan, inputResource.ResultChan)
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
	close(processor.quitChan)
	processor.log.Infof("Stopping TelemetryToArangoProcessor '%s'", processor.name)
}
