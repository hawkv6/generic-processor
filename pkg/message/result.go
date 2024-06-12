package message

import (
	"github.com/hawkv6/generic-processor/pkg/config"
)

type Result interface{ isResult() }
type BaseResultmessage struct{}

func (BaseResultmessage) isResult() {}

type InfluxResultMessage struct {
	BaseResultmessage
	OutputOptions map[string]config.OutputOption
	Results       []InfluxResult
}

type InfluxResult struct {
	Tags  map[string]string
	Value float64
}

type ArangoEventNotificationMessage struct {
	BaseResultmessage
	EventMessages []EventMessage
	Action        string
}
type EventMessage struct {
	Key       string
	Id        string
	TopicType int
}

type ArangoNormalizationMessage struct {
	BaseResultmessage
	Measurement           string
	NormalizationMessages []NormalizationMessage
}

type NormalizationMessage struct {
	Tags   map[string]string
	Fields map[string]float64
}
