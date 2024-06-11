package message

import "github.com/hawkv6/generic-processor/pkg/config"

type Command interface{ isCommand() }
type BaseCommand struct{}

func (BaseCommand) isCommand() {}

type KafkaListeningCommand struct{ BaseCommand }

type InfluxQueryCommand struct {
	BaseCommand
	Measurement    string
	Field          string
	Method         string
	Transformation *config.Transformation
	GroupBy        []string
	Interval       uint
	OutputOptions  map[string]config.OutputOption
}

type ArangoUpdateCommand struct {
	BaseCommand
	Collection string
	Updates    map[string]ArangoUpdate
}

func NewArangoUpdateCommand(collection string) *ArangoUpdateCommand {
	return &ArangoUpdateCommand{
		Collection: collection,
		Updates:    make(map[string]ArangoUpdate),
	}
}

type ArangoUpdate struct {
	Fields []string
	Values []float64
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
