package message

import "github.com/hawkv6/generic-processor/pkg/config"

type Command interface{ isCommand() }
type BaseCommand struct{}

func (BaseCommand) isCommand() {}

type KafkaListeningCommand struct{ BaseCommand }

type InfluxQueryCommand struct {
	BaseCommand
	Measurement   string
	Field         string
	Method        string
	GroupBy       []string
	Interval      uint
	OutputOptions map[string]config.OutputOption
}

type ArangoUpdateCommand struct {
	BaseCommand
	Collection string
	Updates    map[string]ArangoUpdate
}

type ArangoUpdate struct {
	FilterBy map[string]interface{}
	Field    string
	Value    interface{}
	Index    *uint
}

type KafkaUpdateCommand struct {
	BaseCommand
	Updates []KafkaEventMessage
}

type KafkaEventMessage struct {
	TopicType int    `json:"TopicType"`
	Key       string `json:"_key"`
	Id        string `json:"_id"`
	Action    string `json:"action"`
}
