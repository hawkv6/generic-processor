package output

import (
	"fmt"

	"github.com/IBM/sarama"
)

type KafkaProducerMock struct {
	msgChan        chan *sarama.ProducerMessage
	errChan        chan *sarama.ProducerError
	returnCloseErr bool
}

func NewKafkaProducerMock() *KafkaProducerMock {
	return &KafkaProducerMock{
		msgChan:        make(chan *sarama.ProducerMessage, 1),
		errChan:        make(chan *sarama.ProducerError, 1),
		returnCloseErr: false,
	}
}

func (producer *KafkaProducerMock) Close() error {
	if producer.returnCloseErr {
		return fmt.Errorf("Error closing producer")
	}
	return nil
}

func (producer *KafkaProducerMock) Input() chan<- *sarama.ProducerMessage {
	return producer.msgChan
}

func (producer *KafkaProducerMock) Errors() <-chan *sarama.ProducerError {
	return producer.errChan
}
