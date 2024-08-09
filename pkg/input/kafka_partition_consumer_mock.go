package input

import (
	"fmt"

	"github.com/IBM/sarama"
)

type KafkaParitionConsumerMock struct {
	messageChan chan *sarama.ConsumerMessage
	errChan     chan *sarama.ConsumerError
	ReturnError bool
}

func NewKafkaParitionConsumerMock() *KafkaParitionConsumerMock {
	return &KafkaParitionConsumerMock{
		messageChan: make(chan *sarama.ConsumerMessage),
		errChan:     make(chan *sarama.ConsumerError),
		ReturnError: false,
	}
}

func (partitionConsumer *KafkaParitionConsumerMock) Close() error {
	if partitionConsumer.ReturnError {
		return fmt.Errorf("Error closing partition consumer")
	}
	return nil
}

func (partitionConsumer *KafkaParitionConsumerMock) Messages() <-chan *sarama.ConsumerMessage {
	return partitionConsumer.messageChan
}

func (partitionConsumer *KafkaParitionConsumerMock) Errors() <-chan *sarama.ConsumerError {
	return partitionConsumer.errChan
}
