package input

import "github.com/IBM/sarama"

type KafkaConsumer interface {
	ConsumePartition(topic string, partition int32, offset int64) (sarama.PartitionConsumer, error)
	Close() error
}
