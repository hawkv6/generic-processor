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
	FilterBy   map[string]interface{}
	Field      string
	Value      interface{}
	Index      *uint
}
