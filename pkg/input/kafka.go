package input

import (
	"time"

	"github.com/IBM/sarama"
	"github.com/hawkv6/generic-processor/pkg/config"
	"github.com/hawkv6/generic-processor/pkg/logging"
	"github.com/hawkv6/generic-processor/pkg/message"
	"github.com/sirupsen/logrus"
)

type KafkaInput struct {
	log                     *logrus.Entry
	inputConfig             config.KafkaInputConfig
	commandChan             chan message.CommandMessage
	resultChan              chan message.ResultMessage
	quitChan                chan bool
	saramaConfig            *sarama.Config
	saramaConsumer          sarama.Consumer
	saramaPartitionConsumer sarama.PartitionConsumer
}

func NewKafkaInput(config config.KafkaInputConfig, commandChan chan message.CommandMessage, resultChan chan message.ResultMessage) *KafkaInput {
	return &KafkaInput{
		log:         logging.DefaultLogger.WithField("subsystem", Subsystem),
		inputConfig: config,
		commandChan: commandChan,
		resultChan:  resultChan,
		quitChan:    make(chan bool),
	}
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

func (input *KafkaInput) Start() {
	for {
		select {
		case msg := <-input.saramaPartitionConsumer.Messages():
			input.log.Debugln("Received message: ", msg)
			input.resultChan <- message.BaseResultmessage{}
		case <-input.quitChan:
			input.log.Debugln("Stopping Kafka input")
		}
	}
}

func (input *KafkaInput) Stop() {
	input.quitChan <- true
}
