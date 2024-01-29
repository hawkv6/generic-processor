package input

import "github.com/hawkv6/generic-processor/pkg/message"

type InputResource struct {
	Input          Input
	ResultChannel  chan message.ResultMessage
	CommandChannel chan message.Command
}

func NewInputResource() InputResource {
	return InputResource{
		ResultChannel:  make(chan message.ResultMessage),
		CommandChannel: make(chan message.Command),
	}
}
