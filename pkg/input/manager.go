package input

import (
	"fmt"
	"sync"

	"github.com/hawkv6/generic-processor/pkg/config"
	"github.com/hawkv6/generic-processor/pkg/logging"
	"github.com/sirupsen/logrus"
)

type InputManager interface {
	InitInputs() error
	GetInputResources(string) (error, *InputResource)
	StartInputs() error
	StopInputs()
}
type DefaultInputManager struct {
	log            *logrus.Entry
	config         config.Config
	inputResources map[string]InputResource
	wg             sync.WaitGroup
}

func NewDefaultInputManager(config config.Config) *DefaultInputManager {
	return &DefaultInputManager{
		log:            logging.DefaultLogger.WithField("subsystem", Subsystem),
		config:         config,
		inputResources: make(map[string]InputResource),
		wg:             sync.WaitGroup{},
	}
}

func (manager *DefaultInputManager) initInfluxInput(name string, configType config.InfluxInputConfig) error {
	inputResource := NewInputResource()
	input := NewInfluxInput(configType, inputResource.CommandChan, inputResource.ResultChan)
	if err := input.Init(); err != nil {
		return fmt.Errorf("error creating InfluxDB client: %v", err)
	}
	inputResource.Input = input
	manager.inputResources[name] = inputResource
	manager.log.Debugf("Successfully initialized InfluxDB input '%s': ", name)
	return nil
}

func (manager *DefaultInputManager) initKafkaInput(name string, configType config.KafkaInputConfig) error {
	inputResource := NewInputResource()
	input := NewKafkaInput(configType, inputResource.CommandChan, inputResource.ResultChan)
	if err := input.Init(); err != nil {
		return fmt.Errorf("error creating Kafka consumer: %v", err)
	}
	inputResource.Input = input
	manager.inputResources[name] = inputResource
	manager.log.Debugf("Successfully initialized Kafka input '%s': ", name)
	return nil
}

func (manager *DefaultInputManager) InitInputs() error {
	for name, inputConfig := range manager.config.GetInputs() {
		switch inputConfigType := inputConfig.(type) {
		case config.InfluxInputConfig:
			if err := manager.initInfluxInput(name, inputConfigType); err != nil {
				return err
			}
		case config.KafkaInputConfig:
			if err := manager.initKafkaInput(name, inputConfigType); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown input type: %v", inputConfigType)
		}
	}
	return nil
}

func (manager *DefaultInputManager) GetInputResources(name string) (error, *InputResource) {
	input, ok := manager.inputResources[name]
	if !ok {
		return fmt.Errorf("input '%s' not found", name), nil
	}
	return nil, &input
}

func (manager *DefaultInputManager) StartInputs() error {
	manager.log.Infoln("Starting all inputs")
	manager.wg.Add(len(manager.inputResources))
	for _, input := range manager.inputResources {
		go func(input InputResource) {
			defer manager.wg.Done()
			input.Input.Start()
		}(input)
	}
	return nil
}

func (manager *DefaultInputManager) StopInputs() {
	manager.log.Infoln("Stopping all inputs")
	for _, input := range manager.inputResources {
		if err := input.Input.Stop(); err != nil {
			manager.log.Errorln("Error stopping input: ", err)
		}
	}
	manager.wg.Wait()
	manager.log.Infoln("All inputs stopped")
}
