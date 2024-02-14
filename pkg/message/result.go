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
	Value interface{}
}
