package output

import (
	"encoding/json"
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
	log          *logrus.Entry
	config       config.KafkaOutputConfig
	saramaConfig *sarama.Config
	producer     sarama.AsyncProducer
	commandChan  chan message.Command
	resultChan   chan message.Result
	quitChan     chan struct{}
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
	output.saramaConfig = sarama.NewConfig()
	output.saramaConfig.Net.DialTimeout = time.Second * 5
}

func (output *KafkaOutput) Init() error {
	output.createConfig()
	producer, err := sarama.NewAsyncProducer([]string{output.config.Broker}, output.saramaConfig)
	if err != nil {
		output.log.Debugln("Error creating producer: ", err)
		return err
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
	select {
	case output.producer.Input() <- &sarama.ProducerMessage{Topic: output.config.Topic, Key: nil, Value: sarama.StringEncoder(jsonMsg)}:
	case err := <-output.producer.Errors():
		output.log.Errorln("Failed to produce message", err)
	}
}

func (output *KafkaOutput) publishNotificationMessage(msg message.KafkaNormalizationMessage) {
	var enc lineprotocol.Encoder
	enc.SetPrecision(lineprotocol.Nanosecond)
	enc.StartLine(msg.Measurement)

	var keys []string
	for key := range msg.Tags {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		enc.AddTag(key, msg.Tags[key])
	}

	for key, field := range msg.Fields {
		enc.AddField(key, lineprotocol.MustNewValue(float64(field)))
	}
	enc.EndLine(time.Now())
	if err := enc.Err(); err != nil {
		output.log.Errorln("Error encoding line protocol: ", err)
		return
	}

	select {
	case output.producer.Input() <- &sarama.ProducerMessage{Topic: output.config.Topic, Key: nil, Value: sarama.ByteEncoder(enc.Bytes())}:
	case err := <-output.producer.Errors():
		output.log.Errorln("Failed to produce message", err)
	}

}

func (output *KafkaOutput) Start() {
	output.log.Debugln("Starting Kafka output: ", output.config.Name)
	for {
		select {
		case command := <-output.commandChan:
			switch commandType := command.(type) {
			case message.KafkaEventCommand:
				if output.config.Type == "event-notification" {
					output.log.Infof("Publish %d event messages to Kafka\n", len(commandType.Updates))
					for _, msg := range commandType.Updates {
						output.publishMessage(msg)
					}
				}
			case message.KafkaNormalizationCommand:
				if output.config.Type == "normalization" {
					output.log.Infof("Publish %d normalization messages to Kafka\n", len(commandType.Updates))
					for _, msg := range commandType.Updates {
						output.publishNotificationMessage(msg)
					}
				}
			default:
				output.log.Errorf("Unsupported command type: %v", commandType)
			}
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
