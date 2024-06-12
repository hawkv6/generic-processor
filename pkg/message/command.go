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
	Collection      string
	StatisticalData map[string]float64
	Updates         map[string]ArangoUpdate
}

func NewArangoUpdateCommand(collection string) *ArangoUpdateCommand {
	return &ArangoUpdateCommand{
		Collection:      collection,
		Updates:         make(map[string]ArangoUpdate),
		StatisticalData: make(map[string]float64),
	}
}

type ArangoUpdate struct {
	Tags   map[string]string
	Fields []string
	Values []float64 // TODO: Think about introduce map for Fields and Values
}

type KafkaEventCommand struct {
	BaseCommand
	Updates []KafkaEventMessage
}

type KafkaEventMessage struct {
	TopicType int    `json:"TopicType"`
	Key       string `json:"_key"`
	Id        string `json:"_id"`
	Action    string `json:"action"`
}

type KafkaNormalizationCommand struct {
	BaseCommand
	Updates []KafkaNormalizationMessage
}

type KafkaNormalizationMessage struct {
	Measurement string
	Tags        map[string]string
	Fields      map[string]float64
}
