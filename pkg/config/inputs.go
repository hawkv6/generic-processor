package config

type InputConfig interface {
	IsInput()
}

type BaseInputConfig struct {
	Name string `koanf:"name" validate:"required"`
}

func (BaseInputConfig) IsInput() {}

type InfluxInputConfig struct {
	BaseInputConfig
	URL      string `koanf:"url" validate:"required,url"`
	DB       string `koanf:"db" validate:"required"`
	Username string `koanf:"username" validate:"required"`
	Password string `koanf:"password" validate:"required"`
	Timeout  uint   `koanf:"timeout" validate:"required"`
}

type KafkaInputConfig struct {
	BaseInputConfig
	Broker string `koanf:"broker" validate:"required,hostname_port"`
	Topic  string `koanf:"topic" validate:"required"`
}
