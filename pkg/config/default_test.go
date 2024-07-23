package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDefaultConfig(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "TestNewDefaultConfig",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configLocation := "../../tests/config/valid-config.yaml"
			assert.NotNil(t, NewDefaultConfig(configLocation))
		})
	}
}

func TestDefaultConfig_Read(t *testing.T) {
	tests := []struct {
		name           string
		configLocation string
		wantErr        bool
	}{
		{
			name:           "TestDefaultConfig_Read valid file",
			configLocation: "../../tests/config/valid-config.yaml",
			wantErr:        false,
		},
		{
			name:           "TestDefaultConfig_Read empty file location",
			configLocation: "",
			wantErr:        true,
		},
		{
			name:           "TestDefaultConfig_Read json file",
			configLocation: "../../tests/config/invalid-config.json",
			wantErr:        true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewDefaultConfig(tt.configLocation)
			if !tt.wantErr {
				assert.NoError(t, config.Read())
			} else {
				assert.Error(t, config.Read())
			}
		})
	}
}

func TestDefaultConfig_validateInfluxInputConfig(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		db       string
		username string
		password string
		timeout  uint
		wantErr  bool
	}{
		{
			name:     "TestDefaultConfig_validateInfluxInputConfig valid",
			url:      "http://localhost:8086",
			db:       "test",
			username: "test",
			password: "test",
			timeout:  10,
			wantErr:  false,
		},
		{
			name:     "TestDefaultConfig_validateInfluxInputConfig valid",
			url:      "http://localhost:8086",
			db:       "test",
			username: "test",
			password: "test",
			timeout:  0,
			wantErr:  false,
		},
		{
			name:     "TestDefaultConfig_validateInfluxInputConfig invalid",
			url:      "",
			db:       "test",
			username: "test",
			password: "test",
			timeout:  0,
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configLocation := "../../config/valid-config.yaml"
			config := NewDefaultConfig(configLocation)
			influxInput := InfluxInputConfig{
				URL:      tt.url,
				DB:       tt.db,
				Username: tt.username,
				Password: tt.password,
				Timeout:  tt.timeout,
			}
			if tt.wantErr {
				assert.Error(t, config.validateInfluxInputConfig(tt.name, influxInput))
				return
			}
			assert.NoError(t, config.validateInfluxInputConfig(tt.name, influxInput))
		})
	}
}

func TestDefaultConfig_validateInfluxInputs(t *testing.T) {
	tests := []struct {
		name           string
		configLocation string
		wantErr        bool
	}{
		{
			name:           "TestDefaultConfig_validateInfluxInputs success",
			configLocation: "../../tests/config/valid-config.yaml",
			wantErr:        false,
		},
		{
			name:           "TestDefaultConfig_validateInfluxInputs invalid",
			configLocation: "../../tests/config/invalid-influx-input.yaml",
			wantErr:        true,
		},
		{
			name:           "TestDefaultConfig_validateInfluxInputs unmarshall error",
			configLocation: "../../tests/config/marshal-error-config.yaml",
			wantErr:        true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewDefaultConfig(tt.configLocation)
			assert.NoError(t, config.Read())
			if tt.wantErr {
				assert.Error(t, config.validateInfluxInputs())
				return
			}
			assert.NoError(t, config.validateInfluxInputs())
		})
	}
}

func TestDefaultConfig_validateKafkaInputConfig(t *testing.T) {
	tests := []struct {
		name    string
		broker  string
		topic   string
		wantErr bool
	}{
		{
			name:    "TestDefaultConfig_validateKafkaInputConfig valid",
			broker:  "localhost:9092",
			topic:   "test",
			wantErr: false,
		},
		{
			name:    "TestDefaultConfig_validateKafkaInputConfig invalid",
			broker:  "",
			topic:   "test",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configLocation := "../../config/valid-config.yaml"
			config := NewDefaultConfig(configLocation)
			kafkaInput := KafkaInputConfig{
				Broker: tt.broker,
				Topic:  tt.topic,
			}
			if tt.wantErr {
				assert.Error(t, config.validateKafkaInputConfig(tt.name, kafkaInput))
				return
			}
			assert.NoError(t, config.validateKafkaInputConfig(tt.name, kafkaInput))
		})
	}
}

func TestDefaultConfig_validateKafkaInputs(t *testing.T) {
	tests := []struct {
		name           string
		configLocation string
		wantErr        bool
	}{
		{
			name:           "TestDefaultConfig_validateKafkaInputs success",
			configLocation: "../../tests/config/valid-config.yaml",
			wantErr:        false,
		},
		{
			name:           "TestDefaultConfig_validateKafkaInputs invalid",
			configLocation: "../../tests/config/invalid-kafka-input.yaml",
			wantErr:        true,
		},
		{
			name:           "TestDefaultConfig_validateKafkaInputs unmarshall error",
			configLocation: "../../tests/config/marshal-error-config.yaml",
			wantErr:        true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewDefaultConfig(tt.configLocation)
			assert.NoError(t, config.Read())
			if tt.wantErr {
				assert.Error(t, config.validateKafkaInputs())
				return
			}
			assert.NoError(t, config.validateKafkaInputs())
		})
	}
}

func TestDefaultConfig_validateInputs(t *testing.T) {
	tests := []struct {
		name           string
		configLocation string
		wantErr        bool
	}{
		{
			name:           "TestDefaultConfig_validateInputs success",
			configLocation: "../../tests/config/valid-config.yaml",
			wantErr:        false,
		},
		{
			name:           "TestDefaultConfig_validateInputs no inputs defined",
			configLocation: "../../tests/config/empty-config.yaml",
			wantErr:        true,
		},
		{
			name:           "TestDefaultConfig_validateInputs invalid influx input",
			configLocation: "../../tests/config/invalid-influx-input.yaml",
			wantErr:        true,
		},
		{
			name:           "TestDefaultConfig_validateInputs invalid kafka input",
			configLocation: "../../tests/config/invalid-kafka-input.yaml",
			wantErr:        true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewDefaultConfig(tt.configLocation)
			assert.NoError(t, config.Read())
			if tt.wantErr {
				assert.Error(t, config.validateInputs())
				return
			}
			assert.NoError(t, config.validateInputs())
		})
	}
}

func TestDefaultConfig_validateArangoOutputConfig(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		database string
		username string
		password string
		wantErr  bool
	}{
		{
			name:     "TestDefaultConfig_validateArangoOutputConfig valid",
			url:      "http://localhost:8529",
			database: "test",
			username: "test",
			password: "test",
			wantErr:  false,
		},
		{
			name:     "TestDefaultConfig_validateArangoOutputConfig invalid",
			url:      "",
			database: "test",
			username: "test",
			password: "test",
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configLocation := "../../config/valid-config.yaml"
			config := NewDefaultConfig(configLocation)
			arangoOutput := ArangoOutputConfig{
				URL:      tt.url,
				DB:       tt.database,
				Username: tt.username,
				Password: tt.password,
			}
			if tt.wantErr {
				assert.Error(t, config.validateArangoOutputConfig(tt.name, arangoOutput))
				return
			}
			assert.NoError(t, config.validateArangoOutputConfig(tt.name, arangoOutput))
		})
	}
}

func TestDefaultConfig_validateArangoOutputs(t *testing.T) {
	tests := []struct {
		name           string
		configLocation string
		wantErr        bool
	}{
		{
			name:           "TestDefaultConfig_validateArangoOutputs success",
			configLocation: "../../tests/config/valid-config.yaml",
			wantErr:        false,
		},
		{
			name:           "TestDefaultConfig_validateArangoOutputs invalid",
			configLocation: "../../tests/config/invalid-arango-output.yaml",
			wantErr:        true,
		},
		{
			name:           "TestDefaultConfig_validateArangoOutputs unmarshall error",
			configLocation: "../../tests/config/marshal-error-config.yaml",
			wantErr:        true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewDefaultConfig(tt.configLocation)
			assert.NoError(t, config.Read())
			if tt.wantErr {
				assert.Error(t, config.validateArangoOutputs())
				return
			}
			assert.NoError(t, config.validateArangoOutputs())
		})
	}
}

func TestDefaultConfig_validateKafkaOutputConfig(t *testing.T) {
	tests := []struct {
		name       string
		broker     string
		topic      string
		outputType string
		wantErr    bool
	}{
		{
			name:       "TestDefaultConfig_validateKafkaOutputConfig valid",
			broker:     "localhost:9092",
			topic:      "test",
			outputType: "normalization",
			wantErr:    false,
		},
		{
			name:       "TestDefaultConfig_validateKafkaOutputConfig invalid",
			broker:     "",
			topic:      "test",
			outputType: "normalization",
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configLocation := "../../config/valid-config.yaml"
			config := NewDefaultConfig(configLocation)
			kafkaOutput := KafkaOutputConfig{
				Broker: tt.broker,
				Topic:  tt.topic,
				Type:   tt.outputType,
			}
			if tt.wantErr {
				assert.Error(t, config.validateKafkaOutputConfig(tt.name, kafkaOutput))
				return
			}
			assert.NoError(t, config.validateKafkaOutputConfig(tt.name, kafkaOutput))
		})
	}
}

func TestDefaultConfig_validateKafkaOutputs(t *testing.T) {
	tests := []struct {
		name           string
		configLocation string
		wantErr        bool
	}{
		{
			name:           "TestDefaultConfig_validateKafkaOutputs success",
			configLocation: "../../tests/config/valid-config.yaml",
			wantErr:        false,
		},
		{
			name:           "TestDefaultConfig_validateKafkaOutputs invalid",
			configLocation: "../../tests/config/invalid-kafka-output.yaml",
			wantErr:        true,
		},
		{
			name:           "TestDefaultConfig_validateKafkaOutputs unmarshall error",
			configLocation: "../../tests/config/marshal-error-config.yaml",
			wantErr:        true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewDefaultConfig(tt.configLocation)
			assert.NoError(t, config.Read())
			if tt.wantErr {
				assert.Error(t, config.validateKafkaOutputs())
				return
			}
			assert.NoError(t, config.validateKafkaOutputs())
		})
	}
}

func TestDefaultConfig_validateOutputs(t *testing.T) {
	tests := []struct {
		name           string
		configLocation string
		wantErr        bool
	}{
		{
			name:           "TestDefaultConfig_validateOutputs success",
			configLocation: "../../tests/config/valid-config.yaml",
			wantErr:        false,
		},
		{
			name:           "TestDefaultConfig_validateOutputs no outputs defined",
			configLocation: "../../tests/config/empty-config.yaml",
			wantErr:        true,
		},
		{
			name:           "TestDefaultConfig_validateOutputs invalid arango output",
			configLocation: "../../tests/config/invalid-arango-output.yaml",
			wantErr:        true,
		},
		{
			name:           "TestDefaultConfig_validateOutputs invalid kafka output",
			configLocation: "../../tests/config/invalid-kafka-output.yaml",
			wantErr:        true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewDefaultConfig(tt.configLocation)
			assert.NoError(t, config.Read())
			if tt.wantErr {
				assert.Error(t, config.validateOutputs())
				return
			}
			assert.NoError(t, config.validateOutputs())
		})
	}
}

func TestDefaultConfig_validateTelemetryToArangoInputs(t *testing.T) {
	tests := []struct {
		name    string
		inputs  []string
		wantErr bool
	}{
		{
			name:    "TestDefaultConfig_validateTelemetryToArangoInputs valid",
			inputs:  []string{"jalapeno-influx"},
			wantErr: false,
		},
		{
			name:    "TestDefaultConfig_validateTelemetryToArangoInputs invalid",
			inputs:  []string{"influx", "kafka"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configLocation := "../../tests/config/valid-config.yaml"
			config := NewDefaultConfig(configLocation)
			assert.NoError(t, config.Read())
			assert.NoError(t, config.validateInfluxInputs())
			if tt.wantErr {
				assert.Error(t, config.validateTelemetryToArangoInputs(tt.inputs))
				return
			}
			assert.NoError(t, config.validateTelemetryToArangoInputs(tt.inputs))
		})
	}
}

func TestDefaultConfig_validateTelemetryToArangoOutputs(t *testing.T) {
	tests := []struct {
		name    string
		outputs []string
		wantErr bool
	}{
		{
			name:    "TestDefaultConfig_validateTelemetryToArangoOutputs valid",
			outputs: []string{"jalapeno-arango"},
			wantErr: false,
		},
		{
			name:    "TestDefaultConfig_validateTelemetryToArangoOutputs invalid",
			outputs: []string{"arango", "kafka"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configLocation := "../../tests/config/valid-config.yaml"
			config := NewDefaultConfig(configLocation)
			assert.NoError(t, config.Read())
			assert.NoError(t, config.validateArangoOutputs())
			if tt.wantErr {
				assert.Error(t, config.validateTelemetryToArangoOutputs(tt.outputs))
				return
			}
			assert.NoError(t, config.validateTelemetryToArangoOutputs(tt.outputs))
		})
	}
}

func TestDefaultConfig_validateTelemetryToArango(t *testing.T) {
	tests := []struct {
		name            string
		inputs          []string
		outputs         []string
		wantInputErr    bool
		wantOutputErr   bool
		wantMarshallErr bool
		configLocation  string
	}{
		{
			name:            "TestDefaultConfig_validateTelemetryToArango valid",
			inputs:          []string{"jalapeno-influx"},
			outputs:         []string{"jalapeno-arango"},
			wantInputErr:    false,
			wantOutputErr:   false,
			wantMarshallErr: false,
			configLocation:  "../../tests/config/valid-config.yaml",
		},
		{
			name:            "TestDefaultConfig_validateTelemetryToArango invalid inputs",
			inputs:          []string{"influx"},
			outputs:         []string{"jalapeno-arango"},
			wantInputErr:    true,
			wantOutputErr:   false,
			wantMarshallErr: false,
			configLocation:  "../../tests/config/valid-config.yaml",
		},
		{
			name:            "TestDefaultConfig_validateTelemetryToArango invalid outputs",
			inputs:          []string{"jalapeno-influx"},
			outputs:         []string{"arango"},
			wantInputErr:    false,
			wantOutputErr:   true,
			wantMarshallErr: false,
			configLocation:  "../../tests/config/valid-config.yaml",
		},
		{
			name:            "TestDefaultConfig_validateTelemetryToArango marshall error",
			inputs:          []string{"jalapeno-influx"},
			outputs:         []string{"arango"},
			wantInputErr:    false,
			wantOutputErr:   false,
			wantMarshallErr: true,
			configLocation:  "../../tests/config/marshal-error-config.yaml",
		},
		{
			name:            "TestDefaultConfig_validateTelemetryToArango invalid telemetry to arango processor config",
			inputs:          []string{"jalapeno-influx"},
			outputs:         []string{"arango"},
			wantInputErr:    false,
			wantOutputErr:   false,
			wantMarshallErr: true,
			configLocation:  "../../tests/config/invalid-telemetrytoarango-config.yaml",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewDefaultConfig(tt.configLocation)
			assert.NoError(t, config.Read())
			if tt.wantMarshallErr {
				assert.Error(t, config.validateTelemetryToArango())
				return
			}
			if !tt.wantInputErr {
				assert.NoError(t, config.validateInputs())
			}
			if !tt.wantOutputErr {
				assert.NoError(t, config.validateOutputs())
			}
			if tt.wantInputErr || tt.wantOutputErr {
				assert.Error(t, config.validateTelemetryToArango())
				return
			}
			assert.NoError(t, config.validateTelemetryToArango())
		})
	}
}

func TestDefaultConfig_validateProcessors(t *testing.T) {
	tests := []struct {
		name           string
		inputs         []string
		outputs        []string
		wantErr        bool
		configLocation string
	}{
		{
			name:           "TestDefaultConfig_validateTelemetryToArango valid",
			inputs:         []string{"jalapeno-influx"},
			outputs:        []string{"jalapeno-arango"},
			wantErr:        false,
			configLocation: "../../tests/config/valid-config.yaml",
		},
		{
			name:           "TestDefaultConfig_validateTelemetryToArango invalid inputs",
			inputs:         []string{"influx"},
			outputs:        []string{"jalapeno-arango"},
			wantErr:        true,
			configLocation: "../../tests/config/valid-config.yaml",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewDefaultConfig(tt.configLocation)
			assert.NoError(t, config.Read())
			if !tt.wantErr {
				assert.NoError(t, config.validateInputs())
				assert.NoError(t, config.validateOutputs())
			}
			if tt.wantErr {
				assert.Error(t, config.validateProcessors())
				return
			}
			assert.NoError(t, config.validateProcessors())
		})
	}
}

func TestDefaultConfig_Validate(t *testing.T) {
	tests := []struct {
		name           string
		configLocation string
		wantErr        bool
	}{
		{
			name:           "TestDefaultConfig_Validate success",
			configLocation: "../../tests/config/valid-config.yaml",
			wantErr:        false,
		},
		{
			name:           "TestDefaultConfig_Validate invalid influx input",
			configLocation: "../../tests/config/invalid-influx-input.yaml",
			wantErr:        true,
		},
		{
			name:           "TestDefaultConfig_Validate invalid arango output",
			configLocation: "../../tests/config/invalid-arango-output.yaml",
			wantErr:        true,
		},
		{
			name:           "TestDefaultConfig_Validate invalid telemetry to arango",
			configLocation: "../../tests/config/invalid-telemetrytoarango-config.yaml",
			wantErr:        true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewDefaultConfig(tt.configLocation)
			assert.NoError(t, config.Read())
			if tt.wantErr {
				assert.Error(t, config.Validate())
				return
			}
			assert.NoError(t, config.Validate())
		})
	}
}

func TestDefaultConfig_GetInputs(t *testing.T) {
	tests := []struct {
		name           string
		configLocation string
		wantInputKeys  []string
	}{
		{
			name:           "TestDefaultConfig_GetInputs success",
			configLocation: "../../tests/config/valid-config.yaml",
			wantInputKeys:  []string{"jalapeno-influx", "jalapeno-kafka-openconfig"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewDefaultConfig(tt.configLocation)
			assert.NoError(t, config.Read())
			assert.NoError(t, config.Validate())
			inputs := config.GetInputs()
			assert.Len(t, inputs, len(tt.wantInputKeys))
			for _, key := range tt.wantInputKeys {
				assert.Contains(t, inputs, key)
			}
		})
	}
}

func TestDefaultConfig_GetOutputs(t *testing.T) {
	tests := []struct {
		name           string
		configLocation string
		wantOutputKeys []string
	}{
		{
			name:           "TestDefaultConfig_GetOutputs success",
			configLocation: "../../tests/config/valid-config.yaml",
			wantOutputKeys: []string{"jalapeno-arango", "jalapeno-kafka-lslink-events", "jalapeno-kafka-normalization-metrics"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewDefaultConfig(tt.configLocation)
			assert.NoError(t, config.Read())
			assert.NoError(t, config.Validate())
			outputs := config.GetOutputs()
			assert.Len(t, outputs, len(tt.wantOutputKeys))
			for _, key := range tt.wantOutputKeys {
				assert.Contains(t, outputs, key)
			}
		})
	}
}

func TestDefaultConfig_GetProcessors(t *testing.T) {
	tests := []struct {
		name              string
		configLocation    string
		wantProcessorKeys []string
	}{
		{
			name:              "TestDefaultConfig_GetProcessors success",
			configLocation:    "../../tests/config/valid-config.yaml",
			wantProcessorKeys: []string{"LsLink"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewDefaultConfig(tt.configLocation)
			assert.NoError(t, config.Read())
			assert.NoError(t, config.Validate())
			processors := config.GetProcessors()
			assert.Len(t, processors, len(tt.wantProcessorKeys))
			for _, key := range tt.wantProcessorKeys {
				assert.Contains(t, processors, key)
			}
		})
	}
}
