package message

import (
	"github.com/hawkv6/generic-processor/pkg/config"
)

type Result interface{ isResult() }
type BaseResultMessage struct{}

func (BaseResultMessage) isResult() {}

type InfluxResultMessage struct {
	BaseResultMessage
	OutputOptions map[string]config.OutputOption
	Results       []InfluxResult
}

type InfluxResult struct {
	Tags  map[string]string
	Value float64
}

type ArangoEventNotificationMessage struct {
	BaseResultMessage
	EventMessages []EventMessage
	Action        string
}
type EventMessage struct {
	Key       string
	Id        string
	TopicType int
}

type ArangoNormalizationMessage struct {
	BaseResultMessage
	Measurement           string
	NormalizationMessages []NormalizationMessage
}

type NormalizationMessage struct {
	Tags   map[string]string
	Fields map[string]float64
}
