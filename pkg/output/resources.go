package output

import "github.com/hawkv6/generic-processor/pkg/message"

type OutputResource struct {
	Output         Output
	ResultChannel  chan message.ResultMessage
	CommandChannel chan message.CommandMessage
}

func NewOutputResource() OutputResource {
	return OutputResource{
		ResultChannel:  make(chan message.ResultMessage),
		CommandChannel: make(chan message.CommandMessage),
	}
}
