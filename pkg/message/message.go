package message

type ResultMessage interface{ isResultMessage() }
type BaseResultmessage struct{}

func (BaseResultmessage) isResultMessage() {}

type CommandMessage interface{ isCommandMessage() }
type BaseCommandMessage struct{}

func (BaseCommandMessage) isCommandMessage() {}
