package input

import "github.com/IBM/sarama"

type KafkaParitionConsumer interface {
	Close() error
	Messages() <-chan *sarama.ConsumerMessage
	Errors() <-chan *sarama.ConsumerError
}
