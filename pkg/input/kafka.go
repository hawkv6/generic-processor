package input

import (
	"encoding/json"
	"time"

	"github.com/IBM/sarama"
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
	resultChan              chan message.ResultMessage
	quitChan                chan bool
	saramaConfig            *sarama.Config
	saramaConsumer          sarama.Consumer
	saramaPartitionConsumer sarama.PartitionConsumer
	handlers                map[string]func(*message.TelemetryMessage) (message.ResultMessage, error)
}

func NewKafkaInput(config config.KafkaInputConfig, commandChan chan message.Command, resultChan chan message.ResultMessage) *KafkaInput {
	input := &KafkaInput{
		log:         logging.DefaultLogger.WithField("subsystem", Subsystem),
		inputConfig: config,
		commandChan: commandChan,
		resultChan:  resultChan,
		quitChan:    make(chan bool),
	}
	input.handlers = map[string]func(*message.TelemetryMessage) (message.ResultMessage, error){
		"ipv6-addresses": func(msg *message.TelemetryMessage) (message.ResultMessage, error) {
			return input.decodeIpv6Message(msg)
		},
		"oper-state": func(msg *message.TelemetryMessage) (message.ResultMessage, error) {
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
	partitionConsumer, err := input.saramaConsumer.ConsumePartition(input.inputConfig.Topic, 0, sarama.OffsetNewest)
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
	input.log.Debugln("Received JSON message: ", string(msg.Value))
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
		input.log.Debugln("Error decoding IPv6 fields: ", err)
		return nil, err
	}
	ipv6Message := message.IPv6Message{
		TelemetryMessage: *telemetryMessage,
		Fields:           ipv6Fields,
	}
	return &ipv6Message, nil
}

func (input *KafkaInput) decodeInterfaceStatusMessage(telemetryMessage *message.TelemetryMessage) (*message.InterfaceStatusMessage, error) {
	interfaceStatusFields := message.InterfaceStatusFields{}
	if err := mapstructure.Decode(telemetryMessage.Fields, &interfaceStatusFields); err != nil {
		input.log.Debugln("Error decoding Interface Status fields: ", err)
		return nil, err
	}
	interfaceStatusMessage := message.InterfaceStatusMessage{
		TelemetryMessage: *telemetryMessage,
		Fields:           interfaceStatusFields,
	}
	return &interfaceStatusMessage, nil
}

func (input *KafkaInput) processMessage(msg *sarama.ConsumerMessage) {
	input.log.Debugln("Received message: ", string(msg.Value))
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
	input.log.Debugln("Received message: ", message)
	input.resultChan <- message
}

func (input *KafkaInput) StartListening() {
	for {
		msg := <-input.saramaPartitionConsumer.Messages()
		input.processMessage(msg)
	}
}

func (input *KafkaInput) Start() {
	for {
		select {
		case cmd := <-input.commandChan:
			if _, ok := cmd.(message.StartListeningCommand); ok {
				input.StartListening()
			} else {
				input.log.Debugln("Received command: ", cmd)
			}
		case <-input.quitChan:
			input.log.Debugln("Stopping Kafka input")
		}
	}
}

func (input *KafkaInput) Stop() error {
	input.quitChan <- true
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
