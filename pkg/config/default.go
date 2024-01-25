package config

import (
	"flag"
	"fmt"
	"os"

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

func NewDefaultConfig() *DefaultConfig {
	return &DefaultConfig{
		log:           logging.DefaultLogger.WithField("subsystem", Subsystem),
		koanfInstance: koanf.New("."),
		validate:      validator.New(validator.WithRequiredStructEnabled()),
		inputs:        make(map[string]InputConfig),
		outputs:       make(map[string]OutputConfig),
		processors:    make(map[string]ProcessorConfig),
	}
}

func (config *DefaultConfig) init() error {
	configLocation, set := os.LookupEnv("HAWKV6_GENERIC_PROCESSOR_CONFIG")
	if set {
		config.configLocation = configLocation
	} else {
		flag.Parse()
		if configFlag != "" {
			config.configLocation = configFlag
		} else if shortConfigFlag != "" {
			config.configLocation = shortConfigFlag
		} else {
			return fmt.Errorf("config path not set; use -c or --config flag or set HAWKV6_GENERIC_PROCESSOR_CONFIG environment variable")
		}
	}
	return nil
}

func (config *DefaultConfig) Read() error {
	if err := config.init(); err != nil {
		return err
	}
	if err := config.koanfInstance.Load(file.Provider(config.configLocation), yaml.Parser()); err != nil {
		return err
	}
	return nil
}

func (config *DefaultConfig) validateInfluxInputs() error {
	var influxInputs map[string]InfluxInputConfig
	if err := config.koanfInstance.Unmarshal("inputs.influx", &influxInputs); err != nil {
		return err
	}
	for name, influxInput := range influxInputs {
		influxInput.Name = name
		if influxInput.Timeout == 0 {
			influxInput.Timeout = 1
		}
		if err := config.validate.Struct(influxInput); err != nil {
			return err
		}
		config.inputs[name] = influxInput
		config.log.Debugf("Successfully validated influx input '%s'", name)
	}
	config.log.Debugln("Successfully validated all influx inputs")
	return nil
}

func (config *DefaultConfig) validateKafkaInputs() error {
	var kafkaInputs map[string]KafkaInputConfig
	if err := config.koanfInstance.Unmarshal("inputs.kafka", &kafkaInputs); err != nil {
		return err
	}
	for name, kafkaInput := range kafkaInputs {
		kafkaInput.Name = name
		if err := config.validate.Struct(kafkaInput); err != nil {
			return err
		}
		config.inputs[name] = kafkaInput
		config.log.Debugf("Successfully validated kafka input '%s'", name)
	}
	config.log.Debugln("Successfully validated all kafka inputs")
	return nil
}

func (config *DefaultConfig) validateInputs() error {
	if config.koanfInstance.Get("inputs") == nil {
		return fmt.Errorf("No inputs defined in config")
	}
	influxOutputDefined := config.koanfInstance.Get("inputs.influx") != nil
	kafkaOutputDefined := config.koanfInstance.Get("inputs.kafka") != nil
	if !influxOutputDefined && !kafkaOutputDefined {
		return fmt.Errorf("No valid inputs defined in config")
	}
	if influxOutputDefined {
		if err := config.validateInfluxInputs(); err != nil {
			return err
		}
	}
	if kafkaOutputDefined {
		if err := config.validateKafkaInputs(); err != nil {
			return err
		}
	}
	config.log.Debugln("Successfully validated all inputs")
	return nil
}

func (config *DefaultConfig) validateArangoOutputs() error {
	var arangoOutputs map[string]ArangoOutputConfig
	if err := config.koanfInstance.Unmarshal("outputs.arango", &arangoOutputs); err != nil {
		return err
	}
	for name, arangoOutput := range arangoOutputs {
		arangoOutput.Name = name
		if err := config.validate.Struct(arangoOutput); err != nil {
			return err
		}
		config.outputs[name] = arangoOutput
		config.log.Debugf("Successfully validated arango output '%s'", name)
	}
	config.log.Debugln("Successfully validated all arango outputs")
	return nil
}

func (config *DefaultConfig) validateKafkaOutputs() error {
	var kafkaOutputs map[string]KafkaOutputConfig
	if err := config.koanfInstance.Unmarshal("outputs.kafka", &kafkaOutputs); err != nil {
		return err
	}
	for name, kafkaOutput := range kafkaOutputs {
		kafkaOutput.Name = name
		if err := config.validate.Struct(kafkaOutput); err != nil {
			return err
		}
		config.outputs[name] = kafkaOutput
		config.log.Debugf("Successfully validated kafka output '%s'", name)
	}
	config.log.Debugln("Successfully validated all kafka outputs")
	return nil
}

func (config *DefaultConfig) validateOutputs() error {
	if config.koanfInstance.Get("outputs") == nil {
		return fmt.Errorf("No outputs defined in config")
	}
	arangoDefined := config.koanfInstance.Get("outputs.arango") != nil
	kafkaDefined := config.koanfInstance.Get("outputs.kafka") != nil
	if !kafkaDefined && !arangoDefined {
		return fmt.Errorf("No valid outputs defined in config")
	}
	if arangoDefined {
		if err := config.validateArangoOutputs(); err != nil {
			return err
		}
	}
	if kafkaDefined {
		if err := config.validateKafkaOutputs(); err != nil {
			return err
		}
	}
	config.log.Debugln("Successfully validated all outputs")
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
	if err := config.koanfInstance.Unmarshal("processors.TelemetryToArango", &telemetryToArango); err != nil {
		return err
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
		config.log.Debugf("Successfully validated TelemetryToArango processor '%s'", name)
	}

	config.log.Debugln("Successfully validated all TelemetryToArango processors")
	return nil
}

func (config *DefaultConfig) validateProcessors() error {
	if err := config.validateTelemetryToArango(); err != nil {
		return err
	}
	config.log.Debugf("Successfully validated all processors")
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
