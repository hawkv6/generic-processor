package message

type ResultMessage interface{ isResultMessage() }
type BaseResultmessage struct{}

func (BaseResultmessage) isResultMessage() {}

type InfluxResultMessage struct {
	BaseResultmessage
	InfluxCommand
	Tags  map[string]string
	Value interface{}
}
