package input

import (
	"fmt"

	"github.com/IBM/sarama"
)

type KafkaConsumerMock struct {
	returnError bool
}

func NewKafkaConsumerMock() *KafkaConsumerMock {
	return &KafkaConsumerMock{
		returnError: false,
	}
}

func (consumer *KafkaConsumerMock) ConsumePartition(topic string, partition int32, offset int64) (sarama.PartitionConsumer, error) {
	if topic == "" {
		return nil, fmt.Errorf("Error creating partition consumer")
	} else {
		return nil, nil
	}
}
func (consumer *KafkaConsumerMock) Close() error {
	if consumer.returnError {
		return fmt.Errorf("Error closing partition consumer")
	}
	return nil
}
