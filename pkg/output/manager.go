package output

import (
	"fmt"
	"sync"

	"github.com/hawkv6/generic-processor/pkg/config"
	"github.com/hawkv6/generic-processor/pkg/logging"
	"github.com/sirupsen/logrus"
)

type OutputManager interface {
	InitOutputs() error
	GetOutputResource(string) (*OutputResource, error)
	StartOutputs() error
	StopOutputs()
}

type DefaultOutputManager struct {
	log             *logrus.Entry
	config          config.Config
	outputResources map[string]OutputResource
	wg              sync.WaitGroup
}

func NewDefaultOutputManager(config config.Config) *DefaultOutputManager {
	return &DefaultOutputManager{
		log:             logging.DefaultLogger.WithField("subsystem", Subsystem),
		config:          config,
		outputResources: make(map[string]OutputResource),
		wg:              sync.WaitGroup{},
	}
}

func (manager *DefaultOutputManager) initArangoOutput(name string, configType config.ArangoOutputConfig) error {
	outputResource := NewOutputResource()
	output := NewArangoOutput(configType, outputResource.CommandChan, outputResource.ResultChan)
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
	output := NewKafkaOutput(configType, outputResource.CommandChan, outputResource.ResultChan)
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

func (manager *DefaultOutputManager) GetOutputResource(name string) (*OutputResource, error) {
	output, ok := manager.outputResources[name]
	if !ok {
		return nil, fmt.Errorf("output resource '%s' not found", name)
	}
	return &output, nil
}

func (manager *DefaultOutputManager) StartOutputs() error {
	manager.log.Infoln("Starting all outputs")
	manager.wg.Add(len(manager.outputResources))
	for _, output := range manager.outputResources {
		go func(output OutputResource) {
			defer manager.wg.Done()
			output.Output.Start()
		}(output)
	}
	return nil
}

func (manager *DefaultOutputManager) StopOutputs() {
	manager.log.Infoln("Stopping all outputs")
	for _, output := range manager.outputResources {
		if err := output.Output.Stop(); err != nil {
			manager.log.Errorln("Error stopping output: ", err)
		}
	}
	manager.wg.Wait()
	manager.log.Infoln("All output stopped")
}
