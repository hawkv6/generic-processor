package output

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/IBM/sarama"
	"github.com/hawkv6/generic-processor/pkg/config"
	"github.com/hawkv6/generic-processor/pkg/logging"
	"github.com/hawkv6/generic-processor/pkg/message"
	"github.com/influxdata/line-protocol/v2/lineprotocol"
	"github.com/sirupsen/logrus"
)

type KafkaOutput struct {
	log         *logrus.Entry
	config      config.KafkaOutputConfig
	kafkaConfig *sarama.Config
	producer    KafkaProducer
	commandChan chan message.Command
	resultChan  chan message.Result
	quitChan    chan struct{}
}

func NewKafkaOutput(config config.KafkaOutputConfig, commandChan chan message.Command, resultChan chan message.Result) *KafkaOutput {
	return &KafkaOutput{
		log:         logging.DefaultLogger.WithField("subsystem", Subsystem),
		config:      config,
		commandChan: commandChan,
		resultChan:  resultChan,
		quitChan:    make(chan struct{}),
	}
}

func (output *KafkaOutput) createConfig() {
	output.kafkaConfig = sarama.NewConfig()
	output.kafkaConfig.Net.DialTimeout = time.Second * 5
}

func (output *KafkaOutput) Init() error {
	output.createConfig()
	producer, err := sarama.NewAsyncProducer([]string{output.config.Broker}, output.kafkaConfig)
	if err != nil {
		return fmt.Errorf("Error creating producer: %v ", err)
	}
	output.producer = producer
	return nil
}

func (output *KafkaOutput) publishMessage(msg message.KafkaEventMessage) {
	jsonMsg, err := json.Marshal(msg)
	if err != nil {
		output.log.Errorf("Error marshalling message: %v", err)
		return
	}
	output.producer.Input() <- &sarama.ProducerMessage{Topic: output.config.Topic, Key: nil, Value: sarama.StringEncoder(jsonMsg)}
}

func (*KafkaOutput) sortTags(msg message.KafkaNormalizationMessage) []string {
	keys := make([]string, 0, len(msg.Tags))
	for key := range msg.Tags {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func (output *KafkaOutput) encodeToInfluxLineProtocol(msg message.KafkaNormalizationMessage) (lineprotocol.Encoder, bool) {
	var enc lineprotocol.Encoder
	enc.SetPrecision(lineprotocol.Nanosecond)
	enc.StartLine(msg.Measurement)

	for _, key := range output.sortTags(msg) {
		enc.AddTag(key, msg.Tags[key])
	}

	for key, field := range msg.Fields {
		enc.AddField(key, lineprotocol.MustNewValue(float64(field)))
	}
	enc.EndLine(time.Now())
	if err := enc.Err(); err != nil {
		output.log.Errorln("Error encoding line protocol: ", err)
		return lineprotocol.Encoder{}, true
	}
	return enc, false
}

func (output *KafkaOutput) publishNotificationMessage(msg message.KafkaNormalizationMessage) {
	enc, shouldReturn := output.encodeToInfluxLineProtocol(msg)
	if shouldReturn {
		return
	}
	output.producer.Input() <- &sarama.ProducerMessage{Topic: output.config.Topic, Key: nil, Value: sarama.ByteEncoder(enc.Bytes())}
}

func (output *KafkaOutput) publishEventNotifications(commandType message.KafkaEventCommand) {
	if output.config.Type == "event-notification" {
		output.log.Infof("Publish %d event messages to Kafka\n", len(commandType.Updates))
		for _, msg := range commandType.Updates {
			output.publishMessage(msg)
		}
	}
}

func (output *KafkaOutput) publishNormalizationMessages(commandType message.KafkaNormalizationCommand) {
	if output.config.Type == "normalization" {
		output.log.Infof("Publish %d normalization messages to Kafka\n", len(commandType.Updates))
		for _, msg := range commandType.Updates {
			output.publishNotificationMessage(msg)
		}
	}
}

func (output *KafkaOutput) Start() {
	output.log.Debugln("Starting Kafka output: ", output.config.Name)
	for {
		select {
		case command := <-output.commandChan:
			switch commandType := command.(type) {
			case message.KafkaEventCommand:
				output.publishEventNotifications(commandType)
			case message.KafkaNormalizationCommand:
				output.publishNormalizationMessages(commandType)
			default:
				output.log.Errorf("Unsupported command type: %v", commandType)
			}
		case err := <-output.producer.Errors():
			output.log.Errorln("Failed to produce message", err)
		case <-output.quitChan:
			return
		}
	}
}

func (output *KafkaOutput) Stop() error {
	close(output.quitChan)
	if err := output.producer.Close(); err != nil {
		output.log.Errorln("Error closing producer: ", err)
		return err
	}
	return nil
}
