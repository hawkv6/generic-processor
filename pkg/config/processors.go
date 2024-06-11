package config

type ProcessorConfig interface {
	isProcessor()
}

type BaseProcessorConfig struct {
	Name string `koanf:"name" validate:"required"`
}

func (BaseProcessorConfig) isProcessor() {}

type TelemetryToArangoProcessorConfig struct {
	BaseProcessorConfig
	Inputs        []string      `koanf:"inputs" validate:"required"`
	Outputs       []string      `koanf:"outputs" validate:"required"`
	Interval      uint          `koanf:"interval" validate:"required"`
	Normalization Normalization `koanf:"normalization" validate:"required"`
	Modes         []Mode        `koanf:"modes" validate:"required"`
}

type Normalization struct {
	FieldMappings map[string]string `koanf:"field-mappings" validate:"required"`
}

type Mode struct {
	InputOptions  map[string]InputOption  `koanf:"input-options" validate:"required"`
	OutputOptions map[string]OutputOption `koanf:"output-options" validate:"required"`
}

type InputOption struct {
	Name           string          `koanf:"name" validate:"required"`
	Measurement    string          `koanf:"measurement" validate:"required"`
	Field          string          `koanf:"field" validate:"required"`
	Method         string          `koanf:"method" validate:"required"`
	GroupBy        []string        `koanf:"group_by" validate:"required"`
	Transformation *Transformation `koanf:"transformation"`
}
type Transformation struct {
	Operation string `koanf:"operation" validate:"required"`
	Period    uint   `koanf:"period" validate:"required"`
}

type OutputOption struct {
	Name       string   `koanf:"name" validate:"required"`
	Method     string   `koanf:"method" validate:"required"`
	Collection string   `koanf:"collection" validate:"required"`
	FilterBy   []string `koanf:"filter_by" validate:"required"`
	Field      string   `koanf:"field" validate:"required"`
}
