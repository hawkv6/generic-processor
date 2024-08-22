package processor

import (
	"sync"

	"github.com/hawkv6/generic-processor/pkg/config"
	"github.com/hawkv6/generic-processor/pkg/input"
	"github.com/hawkv6/generic-processor/pkg/logging"
	"github.com/hawkv6/generic-processor/pkg/output"
	"github.com/sirupsen/logrus"
)

type ProcessorManager interface {
	Init() error
	StartProcessors()
	StopProcessors()
}

type DefaultProcessorManager struct {
	log           *logrus.Entry
	config        config.Config
	inputManager  input.InputManager
	outputManager output.OutputManager
	processors    map[string]Processor
	wg            sync.WaitGroup
}

func NewDefaultProcessorManager(config config.Config, inputManager input.InputManager, outputManager output.OutputManager) DefaultProcessorManager {
	return DefaultProcessorManager{
		log:           logging.DefaultLogger.WithField("subsystem", Subsystem),
		config:        config,
		inputManager:  inputManager,
		outputManager: outputManager,
		processors:    make(map[string]Processor),
		wg:            sync.WaitGroup{},
	}
}

func (manager *DefaultProcessorManager) getInputResources() (map[string]input.InputResource, error) {
	inputResources := make(map[string]input.InputResource)
	for name := range manager.config.GetInputs() {
		inputResource, err := manager.inputManager.GetInputResources(name)
		if err != nil {
			return nil, err
		}
		inputResources[name] = *inputResource
	}
	return inputResources, nil
}

func (manager *DefaultProcessorManager) getOutputResources() (map[string]output.OutputResource, error) {
	outputResources := make(map[string]output.OutputResource)
	for name := range manager.config.GetOutputs() {
		outputResource, err := manager.outputManager.GetOutputResource(name)
		if err != nil {
			return nil, err
		}
		outputResources[name] = *outputResource
	}
	return outputResources, nil
}

func (manager *DefaultProcessorManager) Init() error {
	for name, processorConfig := range manager.config.GetProcessors() {
		inputResources, err := manager.getInputResources()
		if err != nil {
			return err
		}
		outputResources, err := manager.getOutputResources()
		if err != nil {
			return err
		}
		switch processorConfigType := processorConfig.(type) {
		case config.TelemetryToArangoProcessorConfig:
			processor := NewTelemetryToArangoProcessor(processorConfigType, inputResources, outputResources, NewKafkaOpenConfigProcessor(), NewMinMaxNormalizer())
			manager.processors[name] = processor
		}
	}
	return nil
}

func (manager *DefaultProcessorManager) StartProcessors() {
	manager.log.Infoln("Starting all processors")
	manager.wg.Add(len(manager.processors))
	for _, processor := range manager.processors {
		go func(processor Processor) {
			defer manager.wg.Done()
			processor.Start()
		}(processor)
	}
}

func (manager *DefaultProcessorManager) StopProcessors() {
	manager.log.Infoln("Stopping processors")
	for _, processor := range manager.processors {
		processor.Stop()
	}
	manager.wg.Wait()
	manager.log.Infoln("All processors stopped")
}
