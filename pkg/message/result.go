package message

type ResultMessage interface{ isResultMessage() }
type BaseResultmessage struct{}

func (BaseResultmessage) isResultMessage() {}
