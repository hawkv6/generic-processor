package processor

import (
	"github.com/hawkv6/generic-processor/pkg/config"
	"github.com/hawkv6/generic-processor/pkg/input"
	"github.com/hawkv6/generic-processor/pkg/logging"
	"github.com/hawkv6/generic-processor/pkg/output"
	"github.com/sirupsen/logrus"
)

type TelemetryToArangoProcessor struct {
	log             *logrus.Entry
	name            string
	config          config.TelemetryToArangoProcessorConfig
	inputResources  map[string]input.InputResource
	outputResources map[string]output.OutputResource
}

func NewTelemetryToArangoProcessor(config config.TelemetryToArangoProcessorConfig, inputResources map[string]input.InputResource, outputResources map[string]output.OutputResource) *TelemetryToArangoProcessor {
	return &TelemetryToArangoProcessor{
		log:             logging.DefaultLogger.WithField("subsystem", Subsystem),
		name:            config.Name,
		config:          config,
		inputResources:  inputResources,
		outputResources: outputResources,
	}
}

func (processor *TelemetryToArangoProcessor) Init() error {
	processor.log.Infof("Initializing TelemetryToArangoProcessor '%s'", processor.name)
	return nil
}

func (processor *TelemetryToArangoProcessor) Start() {
}

func (processor *TelemetryToArangoProcessor) Stop() {
}
