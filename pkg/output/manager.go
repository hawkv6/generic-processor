package output

import (
	"fmt"

	"github.com/hawkv6/generic-processor/pkg/config"
	"github.com/hawkv6/generic-processor/pkg/logging"
	"github.com/sirupsen/logrus"
)

type OutputManager interface {
	InitOutputs() error
	GetOutputResource(string) (error, *OutputResource)
}

type DefaultOutputManager struct {
	log             *logrus.Entry
	config          config.Config
	outputResources map[string]OutputResource
}

func NewDefaultOutputManager(config config.Config) *DefaultOutputManager {
	return &DefaultOutputManager{
		log:             logging.DefaultLogger.WithField("subsystem", Subsystem),
		config:          config,
		outputResources: make(map[string]OutputResource),
	}
}

func (manager *DefaultOutputManager) initArangoOutput(name string, configType config.ArangoOutputConfig) error {
	outputResource := NewOutputResource()
	output := NewArangoOutput(configType)
	if err := output.Init(); err != nil {
		return err
	}
	outputResource.Output = output
	manager.outputResources[name] = outputResource
	manager.log.Debugf("Successfully initialized ArangoDB output '%s': ", name)
	return nil
}

func (manager *DefaultOutputManager) initKafkaOutput(name string, configType config.KafkaOutputConfig) error {
	outputResource := NewOutputResource()
	output := NewKafkaOutput(configType)
	if err := output.Init(); err != nil {
		return err
	}
	outputResource.Output = output
	manager.outputResources[name] = outputResource
	manager.log.Debugf("Successfully initialized Kafka output '%s': ", name)
	return nil
}

func (manager *DefaultOutputManager) InitOutputs() error {
	for name, outputConfig := range manager.config.GetOutputs() {
		switch outputConfigType := outputConfig.(type) {
		case config.ArangoOutputConfig:
			if err := manager.initArangoOutput(name, outputConfigType); err != nil {
				return err
			}
		case config.KafkaOutputConfig:
			if err := manager.initKafkaOutput(name, outputConfigType); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown output type: %v", outputConfigType)

		}
	}
	return nil
}

func (manager *DefaultOutputManager) GetOutputResource(name string) (error, *OutputResource) {
	output, ok := manager.outputResources[name]
	if !ok {
		return fmt.Errorf("output resource '%s' not found", name), nil
	}
	return nil, &output
}
