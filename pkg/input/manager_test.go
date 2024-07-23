package input

import (
	"testing"

	"github.com/hawkv6/generic-processor/pkg/config"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNewDefaultInputManager(t *testing.T) {
	tests := []struct {
		name string
	}{{

		name: "TestNewDefaultInputManager",
	},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := config.NewDefaultConfig("../../tests/config/valid-config.yaml")
			got := NewDefaultInputManager(config)
			if got == nil {
				t.Errorf("NewDefaultInputManager() = %v, want %v", got, nil)
			}
		})
	}
}

func TestDefaultInputManager_initInfluxInput(t *testing.T) {
	tests := []struct {
		name string
	}{{

		name: "TestDefaultInputManager_initInfluxInput connection error ",
	},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := config.InfluxInputConfig{
				URL:      "http://localhost:8086",
				Username: "admin",
				Password: "admin",
				DB:       "test",
				Timeout:  10,
			}
			manager := NewDefaultInputManager(nil)
			assert.Error(t, manager.initInfluxInput("test", config))
		})
	}
}

func TestDefaultInputManager_initKafkaInput(t *testing.T) {
	tests := []struct {
		name string
	}{{

		name: "TestDefaultInputManager_initKafkaInput",
	},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := config.KafkaInputConfig{
				Broker: "localhost:9092",
				Topic:  "test",
			}
			manager := NewDefaultInputManager(nil)
			assert.Error(t, manager.initKafkaInput("test", config))
		})
	}
}

func TestDefaultInputManager_InitInputs(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		// {

		// 	name:    "TestDefaultInputManager_InitInputs empty config",
		// 	wantErr: false,
		// },
		{

			name:    "TestDefaultInputManager_InitInputs connection error",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := config.NewDefaultConfig("../../tests/config/valid-config.yaml")
			if tt.wantErr {
				assert.NoError(t, config.Read())
				assert.NoError(t, config.Validate())
			}
			manager := NewDefaultInputManager(config)
			if tt.wantErr {
				assert.Error(t, manager.InitInputs())
				return
			}
			assert.NoError(t, manager.InitInputs())
		})
	}
}

func TestDefaultInputManager_GetInputResources(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{

			name:    "TestDefaultInputManager_GetInputResources success",
			wantErr: false,
		},
		{
			name:    "TestDefaultInputManager_GetInputResources error",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := config.NewDefaultConfig("../../tests/config/valid-config.yaml")
			manager := NewDefaultInputManager(config)
			if tt.wantErr {
				input, err := manager.GetInputResources("test")
				assert.Nil(t, input)
				assert.Error(t, err)
				return
			}
			manager.inputResources["test"] = NewInputResource()
			input, err := manager.GetInputResources("test")
			assert.NotNil(t, input)
			assert.NoError(t, err)
		})
	}
}

func TestDefaultInputManager_StartInputs(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "TestDefaultInputManager_StartInputs success",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := config.NewDefaultConfig("../../tests/config/valid-config.yaml")
			manager := NewDefaultInputManager(config)
			input := NewMockInput(gomock.NewController(t))
			input.EXPECT().Start().Return()
			inputResource := NewInputResource()
			inputResource.Input = input
			manager.inputResources["test"] = inputResource
			manager.StartInputs()
			manager.wg.Wait()
		})
	}
}

func TestDefaultInputManager_StopInputs(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "TestDefaultInputManager_StopInputs success",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := config.NewDefaultConfig("../../tests/config/valid-config.yaml")
			manager := NewDefaultInputManager(config)
			input := NewMockInput(gomock.NewController(t))
			input.EXPECT().Stop().Return(nil)
			inputResource := NewInputResource()
			inputResource.Input = input
			manager.inputResources["test"] = inputResource
			manager.StopInputs()
		})
	}
}
