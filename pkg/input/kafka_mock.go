package input

import (
	"fmt"

	"github.com/IBM/sarama"
)

type KafkaConsumer interface {

	// ConsumePartition creates a PartitionConsumer on the given topic/partition with
	// the given offset. It will return an error if this Consumer is already consuming
	// on the given topic/partition. Offset can be a literal offset, or OffsetNewest
	// or OffsetOldest
	ConsumePartition(topic string, partition int32, offset int64) (sarama.PartitionConsumer, error)

	// Close shuts down the consumer. It must be called after all child
	// PartitionConsumers have already been closed.
	Close() error
}

type KafkaConsumerMock struct {
}

func NewKafkaConsumerMock() *KafkaConsumerMock {
	return &KafkaConsumerMock{}
}

func (consumer *KafkaConsumerMock) ConsumePartition(topic string, partition int32, offset int64) (sarama.PartitionConsumer, error) {
	if topic == "" {
		return nil, fmt.Errorf("Error creating partition consumer")
	} else {
		return nil, nil
	}
}
func (consumer *KafkaConsumerMock) Close() error {
	return nil
}

type KafkaParitionConsumer interface {
	// Close stops the PartitionConsumer from fetching messages. It will initiate a shutdown just like AsyncClose, drain
	// the Messages channel, harvest any errors & return them to the caller. Note that if you are continuing to service
	// the Messages channel when this function is called, you will be competing with Close for messages; consider
	// calling AsyncClose, instead. It is required to call this function (or AsyncClose) before a consumer object passes
	// out of scope, as it will otherwise leak memory. You must call this before calling Close on the underlying client.
	Close() error

	// Messages returns the read channel for the messages that are returned by
	// the broker.
	Messages() <-chan *sarama.ConsumerMessage

	// Errors returns a read channel of errors that occurred during consuming, if
	// enabled. By default, errors are logged and not returned over this channel.
	// If you want to implement any custom error handling, set your config's
	// Consumer.Return.Errors setting to true, and read from this channel.
	Errors() <-chan *sarama.ConsumerError
}

type KafkaParitionConsumerMock struct {
	messageChan chan *sarama.ConsumerMessage
	errChan     chan *sarama.ConsumerError
}

func NewKafkaParitionConsumerMock() *KafkaParitionConsumerMock {
	return &KafkaParitionConsumerMock{
		messageChan: make(chan *sarama.ConsumerMessage),
		errChan:     make(chan *sarama.ConsumerError),
	}
}

func (partitionConsumer *KafkaParitionConsumerMock) Close() error {
	return nil
}

func (partitionConsumer *KafkaParitionConsumerMock) Messages() <-chan *sarama.ConsumerMessage {
	return partitionConsumer.messageChan
}

func (partitionConsumer *KafkaParitionConsumerMock) Errors() <-chan *sarama.ConsumerError {
	return partitionConsumer.errChan
}
