package output

import (
	"encoding/json"
	"time"

	"github.com/IBM/sarama"
	"github.com/hawkv6/generic-processor/pkg/config"
	"github.com/hawkv6/generic-processor/pkg/logging"
	"github.com/hawkv6/generic-processor/pkg/message"
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

func (output *KafkaOutput) Start() {
	output.log.Debugln("Starting Kafka output: ", output.config.Name)
	for {
		select {
		// TODO add a new command to publish dat to Kafka so that Telegraf can be used to send data to InfluxDB
		case command := <-output.commandChan:
			switch commandType := command.(type) {
			case message.KafkaUpdateCommand:
				output.log.Infof("Publish %d event messages to Kafka\n", len(commandType.Updates))
				for _, msg := range commandType.Updates {
					output.publishMessage(msg)
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
