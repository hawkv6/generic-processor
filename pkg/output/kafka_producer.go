package output

import "github.com/IBM/sarama"

type KafkaProducer interface {
	Close() error
	Input() chan<- *sarama.ProducerMessage
	Errors() <-chan *sarama.ProducerError
}
