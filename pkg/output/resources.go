package output

import "github.com/hawkv6/generic-processor/pkg/message"

type OutputResource struct {
	Output      Output
	ResultChan  chan message.Result
	CommandChan chan message.Command
}

func NewOutputResource() OutputResource {
	return OutputResource{
		ResultChan:  make(chan message.Result),
		CommandChan: make(chan message.Command),
	}
}
