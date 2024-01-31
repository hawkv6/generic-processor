package input

import "github.com/hawkv6/generic-processor/pkg/message"

type InputResource struct {
	Input       Input
	ResultChan  chan message.Result
	CommandChan chan message.Command
}

func NewInputResource() InputResource {
	return InputResource{
		ResultChan:  make(chan message.Result),
		CommandChan: make(chan message.Command),
	}
}
