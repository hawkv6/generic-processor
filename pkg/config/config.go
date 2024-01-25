package config

const Subsystem = "config"

type Config interface {
	Read() error
	Validate() error
	GetInputs() map[string]InputConfig
	GetOutputs() map[string]OutputConfig
	GetProcessors() map[string]ProcessorConfig
}
