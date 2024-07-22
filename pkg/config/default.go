package config

import (
	"fmt"

	"github.com/hawkv6/generic-processor/pkg/logging"
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/sirupsen/logrus"

	"github.com/go-playground/validator/v10"
)

type DefaultConfig struct {
	log            *logrus.Entry
	koanfInstance  *koanf.Koanf
	validate       *validator.Validate
	configLocation string
	inputs         map[string]InputConfig
	outputs        map[string]OutputConfig
	processors     map[string]ProcessorConfig
}

func NewDefaultConfig(configLocation string) *DefaultConfig {
	return &DefaultConfig{
		log:            logging.DefaultLogger.WithField("subsystem", Subsystem),
		koanfInstance:  koanf.New("."),
		validate:       validator.New(validator.WithRequiredStructEnabled()),
		inputs:         make(map[string]InputConfig),
		outputs:        make(map[string]OutputConfig),
		processors:     make(map[string]ProcessorConfig),
		configLocation: configLocation,
	}
}

func (config *DefaultConfig) Read() error {
	if config.configLocation == "" {
		return fmt.Errorf("config path not set; use -c or --config flag or set HAWKV6_GENERIC_PROCESSOR_CONFIG environment variable")
	}
	if err := config.koanfInstance.Load(file.Provider(config.configLocation), yaml.Parser()); err != nil {
		return err
	}
	return nil
}

func (config *DefaultConfig) validateInfluxInputConfig(name string, influxInput InfluxInputConfig) error {
	config.log.Infof("Validating influx input '%s'", name)
	influxInput.Name = name
	if influxInput.Timeout == 0 {
		influxInput.Timeout = 1
	}
	if err := config.validate.Struct(influxInput); err != nil {
		return err
	}
	config.inputs[name] = influxInput
	config.log.Infof("Successfully validated influx input '%s'", name)
	return nil
}

func (config *DefaultConfig) validateInfluxInputs() error {
	var influxInputs map[string]InfluxInputConfig
	err := config.koanfInstance.Unmarshal("inputs.influx", &influxInputs)
	if err != nil || influxInputs == nil {
		return fmt.Errorf("Error marshalling influx inputs: %v", err)
	}
	for name, influxInput := range influxInputs {
		err := config.validateInfluxInputConfig(name, influxInput)
		if err != nil {
			return err
		}
	}
	config.log.Infoln("Successfully validated all influx inputs")
	return nil
}

func (config *DefaultConfig) validateKafkaInputConfig(name string, kafkaInput KafkaInputConfig) error {
	config.log.Infof("Validating kafka input '%s'", name)
	kafkaInput.Name = name
	if err := config.validate.Struct(kafkaInput); err != nil {
		return err
	}
	config.inputs[name] = kafkaInput
	config.log.Infof("Successfully validated kafka input '%s'", name)
	return nil
}

func (config *DefaultConfig) validateKafkaInputs() error {
	var kafkaInputs map[string]KafkaInputConfig
	err := config.koanfInstance.Unmarshal("inputs.kafka", &kafkaInputs)
	if err != nil || kafkaInputs == nil {
		return fmt.Errorf("Error marshalling kafka inputs: %v", err)
	}
	for name, kafkaInput := range kafkaInputs {
		err := config.validateKafkaInputConfig(name, kafkaInput)
		if err != nil {
			return err
		}
	}
	config.log.Infof("Successfully validated all kafka inputs")
	return nil
}

func (config *DefaultConfig) validateInputs() error {
	influxInputDefined := config.koanfInstance.Get("inputs.influx") != nil
	kafkaInputDefined := config.koanfInstance.Get("inputs.kafka") != nil
	if !influxInputDefined && !kafkaInputDefined {
		return fmt.Errorf("No valid inputs defined in config")
	}
	if influxInputDefined {
		if err := config.validateInfluxInputs(); err != nil {
			return err
		}
	}
	if kafkaInputDefined {
		if err := config.validateKafkaInputs(); err != nil {
			return err
		}
	}
	config.log.Infoln("Successfully validated all inputs")
	return nil
}

func (config *DefaultConfig) validateArangoOutputConfig(name string, arangoOutput ArangoOutputConfig) error {
	config.log.Infof("Validating arango output '%s'", name)
	arangoOutput.Name = name
	if err := config.validate.Struct(arangoOutput); err != nil {
		return err
	}
	config.outputs[name] = arangoOutput
	config.log.Infof("Successfully validated arango output '%s'", name)
	return nil
}

func (config *DefaultConfig) validateArangoOutputs() error {
	var arangoOutputs map[string]ArangoOutputConfig
	err := config.koanfInstance.Unmarshal("outputs.arango", &arangoOutputs)
	if err != nil || arangoOutputs == nil {
		return fmt.Errorf("Error marshalling arango outputs: %v", err)
	}
	for name, arangoOutput := range arangoOutputs {
		err := config.validateArangoOutputConfig(name, arangoOutput)
		if err != nil {
			return err
		}
	}
	config.log.Infoln("Successfully validated all arango outputs")
	return nil
}

func (config *DefaultConfig) validateKafkaOutputConfig(name string, kafkaOutput KafkaOutputConfig) error {
	config.log.Infof("Validating kafka output '%s'", name)
	kafkaOutput.Name = name
	if err := config.validate.Struct(kafkaOutput); err != nil {
		return err
	}
	config.outputs[name] = kafkaOutput
	config.log.Infof("Successfully validated kafka output '%s'", name)
	return nil
}
func (config *DefaultConfig) validateKafkaOutputs() error {
	var kafkaOutputs map[string]KafkaOutputConfig
	err := config.koanfInstance.Unmarshal("outputs.kafka", &kafkaOutputs)
	if err != nil || kafkaOutputs == nil {
		return fmt.Errorf("Error marshalling kafka outputs: %v", err)
	}
	for name, kafkaOutput := range kafkaOutputs {
		err := config.validateKafkaOutputConfig(name, kafkaOutput)
		if err != nil {
			return err
		}
	}
	config.log.Infoln("Successfully validated all kafka outputs")
	return nil
}

func (config *DefaultConfig) validateOutputs() error {
	arangoOutputDefined := config.koanfInstance.Get("outputs.arango") != nil
	kafkaOutputDefined := config.koanfInstance.Get("outputs.kafka") != nil
	if !kafkaOutputDefined && !arangoOutputDefined {
		return fmt.Errorf("No valid outputs defined in config")
	}
	if arangoOutputDefined {
		if err := config.validateArangoOutputs(); err != nil {
			return err
		}
	}
	if kafkaOutputDefined {
		if err := config.validateKafkaOutputs(); err != nil {
			return err
		}
	}
	config.log.Infoln("Successfully validated all outputs")
	return nil
}

func (config *DefaultConfig) validateTelemetryToArangoInputs(inputs []string) error {
	for _, input := range inputs {
		if _, ok := config.inputs[input]; !ok {
			return fmt.Errorf("input '%s' not defined", input)
		}
	}
	return nil
}

func (config *DefaultConfig) validateTelemetryToArangoOutputs(outputs []string) error {
	for _, output := range outputs {
		if _, ok := config.outputs[output]; !ok {
			return fmt.Errorf("output '%s' not defined", output)
		}
	}
	return nil
}

func (config *DefaultConfig) validateTelemetryToArango() error {
	var telemetryToArango map[string]TelemetryToArangoProcessorConfig
	err := config.koanfInstance.Unmarshal("processors.TelemetryToArango", &telemetryToArango)
	if err != nil || telemetryToArango == nil {
		return fmt.Errorf("Error marshalling TelemetryToArango processors: %v", err)
	}
	for name, processor := range telemetryToArango {
		processor.Name = name
		if err := config.validate.Struct(processor); err != nil {
			return err
		}
		if err := config.validateTelemetryToArangoInputs(processor.Inputs); err != nil {
			return err
		}
		if err := config.validateTelemetryToArangoOutputs(processor.Outputs); err != nil {
			return err
		}
		config.processors[name] = processor
		config.log.Infof("Successfully validated TelemetryToArango processor '%s'", name)
	}

	config.log.Infoln("Successfully validated all TelemetryToArango processors")
	return nil
}

func (config *DefaultConfig) validateProcessors() error {
	if err := config.validateTelemetryToArango(); err != nil {
		return err
	}
	config.log.Infof("Successfully validated all processors")
	return nil
}

func (config *DefaultConfig) Validate() error {
	if err := config.validateInputs(); err != nil {
		return err
	}
	if err := config.validateOutputs(); err != nil {
		return err
	}
	if err := config.validateProcessors(); err != nil {
		return err
	}
	return nil

}

func (config *DefaultConfig) GetInputs() map[string]InputConfig {
	return config.inputs
}

func (config *DefaultConfig) GetOutputs() map[string]OutputConfig {
	return config.outputs
}

func (config *DefaultConfig) GetProcessors() map[string]ProcessorConfig {
	return config.processors
}
