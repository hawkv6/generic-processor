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
	GetInputResources(string) (*InputResource, error)
	StartInputs()
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
	input, err := NewInfluxInput(configType, inputResource.CommandChan, inputResource.ResultChan)
	if err != nil {
		return fmt.Errorf("error creating InfluxDB client: %v", err)
	}
	if err := input.Init(); err != nil {
		return fmt.Errorf("error initializing InfluxDB client: %v", err)
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

func (manager *DefaultInputManager) GetInputResources(name string) (*InputResource, error) {
	input, ok := manager.inputResources[name]
	if !ok {
		return nil, fmt.Errorf("input '%s' not found", name)
	}
	return &input, nil
}

func (manager *DefaultInputManager) StartInputs() {
	manager.log.Infoln("Starting all inputs")
	manager.wg.Add(len(manager.inputResources))
	for _, input := range manager.inputResources {
		go func(input InputResource) {
			defer manager.wg.Done()
			input.Input.Start()
		}(input)
	}
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
