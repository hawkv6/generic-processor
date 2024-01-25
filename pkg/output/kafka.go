package output

import (
	"time"

	"github.com/IBM/sarama"
	"github.com/hawkv6/generic-processor/pkg/config"
	"github.com/hawkv6/generic-processor/pkg/logging"
	"github.com/sirupsen/logrus"
)

type KafkaOutput struct {
	log          *logrus.Entry
	config       config.KafkaOutputConfig
	saramaConfig *sarama.Config
	producer     sarama.AsyncProducer
}

func NewKafkaOutput(config config.KafkaOutputConfig) *KafkaOutput {
	return &KafkaOutput{
		log:    logging.DefaultLogger.WithField("subsystem", Subsystem),
		config: config,
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

func (output *KafkaOutput) Start() {}

func (output *KafkaOutput) Stop() {}
