package config

type OutputConfig interface {
	IsOutput()
}

type BaseOutputConfig struct {
	Name string `koanf:"name" validate:"required"`
}

func (BaseOutputConfig) IsOutput() {}

type ArangoOutputConfig struct {
	BaseOutputConfig
	URL      string `koanf:"url" validate:"required,url"`
	DB       string `koanf:"database" validate:"required"`
	Username string `koanf:"username" validate:"required"`
	Password string `koanf:"password" validate:"required"`
}

type KafkaOutputConfig struct {
	BaseOutputConfig
	Broker string `koanf:"broker" validate:"required,hostname_port"`
	Topic  string `koanf:"topic" validate:"required"`
}
