package input

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/IBM/sarama"
	"github.com/go-playground/validator/v10"
	"github.com/hawkv6/generic-processor/pkg/config"
	"github.com/hawkv6/generic-processor/pkg/logging"
	"github.com/hawkv6/generic-processor/pkg/message"
	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
)

type KafkaInput struct {
	log                     *logrus.Entry
	inputConfig             config.KafkaInputConfig
	commandChan             chan message.Command
	resultChan              chan message.Result
	quitChan                chan struct{}
	saramaConfig            *sarama.Config
	saramaConsumer          KafkaConsumer
	saramaPartitionConsumer KafkaParitionConsumer
	handlers                map[string]func(*message.TelemetryMessage) (message.Result, error)
	wg                      sync.WaitGroup
}

func NewKafkaInput(config config.KafkaInputConfig, commandChan chan message.Command, resultChan chan message.Result) *KafkaInput {
	input := &KafkaInput{
		log:         logging.DefaultLogger.WithField("subsystem", Subsystem),
		inputConfig: config,
		commandChan: commandChan,
		resultChan:  resultChan,
		quitChan:    make(chan struct{}),
		wg:          sync.WaitGroup{},
	}
	input.handlers = map[string]func(*message.TelemetryMessage) (message.Result, error){
		"ipv6-addresses": func(msg *message.TelemetryMessage) (message.Result, error) {
			return input.decodeIpv6Message(msg)
		},
		"oper-state": func(msg *message.TelemetryMessage) (message.Result, error) {
			return input.decodeInterfaceStatusMessage(msg)
		},
	}
	return input
}
func (input *KafkaInput) createConfig() {
	input.saramaConfig = sarama.NewConfig()
	input.saramaConfig.Net.DialTimeout = time.Second * 5
}

func (input *KafkaInput) createConsumer() error {
	input.createConfig()
	saramaConsumer, err := sarama.NewConsumer([]string{input.inputConfig.Broker}, input.saramaConfig)
	if err != nil {
		input.log.Debugln("Error creating consumer: ", err)
		return err
	}
	input.log.Debugln("Successfully created Kafka consumer for broker: ", input.inputConfig.Broker)
	input.saramaConsumer = saramaConsumer
	return nil
}
func (input *KafkaInput) createParitionConsumer() error {
	partitionConsumer, err := input.saramaConsumer.ConsumePartition(input.inputConfig.Topic, 0, sarama.OffsetOldest)
	if err != nil {
		input.log.Debugln("Error partition consumer: ", err)
		return err
	}
	input.saramaPartitionConsumer = partitionConsumer
	input.log.Debugln("Successfully created Kafka partition consumer for topic: ", input.inputConfig.Topic)
	return nil
}

func (input *KafkaInput) Init() error {
	if err := input.createConsumer(); err != nil {
		return err
	}
	if err := input.createParitionConsumer(); err != nil {
		return err
	}
	return nil
}

func (input *KafkaInput) UnmarshalTelemetryMessage(msg *sarama.ConsumerMessage) (*message.TelemetryMessage, error) {
	input.log.Debugln("Received JSON message from Kafka, unmarshalling...")
	var telemetryMessage message.TelemetryMessage
	if err := json.Unmarshal([]byte(msg.Value), &telemetryMessage); err != nil {
		input.log.Debugln("Error unmarshalling message: ", err)
		return nil, err
	}
	return &telemetryMessage, nil
}

func (input *KafkaInput) decodeIpv6Message(telemetryMessage *message.TelemetryMessage) (*message.IPv6Message, error) {
	ipv6Fields := message.IPv6Fields{}
	if err := mapstructure.Decode(telemetryMessage.Fields, &ipv6Fields); err != nil {
		return nil, err
	}
	if err := validator.New().Struct(ipv6Fields); err != nil {
		return nil, err
	}
	ipv6Message := message.IPv6Message{
		TelemetryMessage: *telemetryMessage,
		Fields:           ipv6Fields,
	}
	input.log.Debugln("Decoded IPv6 message: ", ipv6Message.Fields.IPv6)
	return &ipv6Message, nil
}

func (input *KafkaInput) decodeInterfaceStatusMessage(telemetryMessage *message.TelemetryMessage) (*message.InterfaceStatusMessage, error) {
	intStatusFields := message.InterfaceStatusFields{}
	if err := mapstructure.Decode(telemetryMessage.Fields, &intStatusFields); err != nil {
		return nil, err
	}
	if err := validator.New().Struct(intStatusFields); err != nil {
		return nil, err
	}
	intStatusMsg := message.InterfaceStatusMessage{
		TelemetryMessage: *telemetryMessage,
		Fields:           intStatusFields,
	}
	input.log.Debugf("Decoded Interface Status message: Router: %s, Interface: %s, AdminStatus %s, OperStatus %s", telemetryMessage.Tags.Source, intStatusMsg.Fields.Name, intStatusMsg.Fields.AdminStatus, intStatusMsg.Fields.OperStatus)
	return &intStatusMsg, nil
}

func (input *KafkaInput) processMessage(msg *sarama.ConsumerMessage) {
	telemetryMessage, err := input.UnmarshalTelemetryMessage(msg)
	if err != nil {
		input.log.Debugln("Error unmarshalling message: ", err)
		return
	}
	handler, ok := input.handlers[telemetryMessage.Name]
	if !ok {
		input.log.Debugln("Unknown message type: ", telemetryMessage.Name)
		return
	}
	message, err := handler(telemetryMessage)
	if err != nil {
		input.log.Debugln("Error decoding message: ", err)
		return
	}
	input.resultChan <- message
}

func (input *KafkaInput) StartListening() {
	for {
		select {
		case msg := <-input.saramaPartitionConsumer.Messages():
			input.processMessage(msg)
		case <-input.quitChan:
			input.log.Infof("Stopping Kafka input '%s'", input.inputConfig.Name)
			return
		}
	}
}

func (input *KafkaInput) Start() {
	input.log.Infof("Starting Kafka input '%s'", input.inputConfig.Name)
	cmd := <-input.commandChan
	if _, ok := cmd.(message.KafkaListeningCommand); ok {
		input.StartListening()
	} else {
		input.log.Debugln("Received unknown command: ", cmd)
	}
}

func (input *KafkaInput) Stop() error {
	close(input.quitChan)
	if err := input.saramaPartitionConsumer.Close(); err != nil {
		input.log.Errorln("Error closing partition consumer: ", err)
		return err
	}
	if err := input.saramaConsumer.Close(); err != nil {
		input.log.Errorln("Error closing consumer: ", err)
		return err
	}
	return nil
}
