package output

import (
	"errors"
	"testing"

	"github.com/hawkv6/generic-processor/pkg/config"
	"github.com/stretchr/testify/assert"
	gomock "go.uber.org/mock/gomock"
)

func TestNewDefaultOutputManager(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "TestNewDefaultOutputManager",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotNil(t, NewDefaultOutputManager(nil))
		})
	}
}

func TestDefaultOutputManager_initArangoOutput(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "TestDefaultOutputManager_initArangoOutput",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewDefaultOutputManager(nil)
			config := config.ArangoOutputConfig{
				URL:      "http://localhost:8529",
				Username: "admin",
				Password: "admin",
				DB:       "test",
			}
			assert.NoError(t, manager.initArangoOutput("test", config))
		})
	}
}

func TestDefaultOutputManager_initKafkaOutput(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "TestDefaultOutputManager_initKafkaOutput",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewDefaultOutputManager(nil)
			config := config.KafkaOutputConfig{
				Broker: "localhost:9092",
				Topic:  "test",
			}
			assert.Error(t, manager.initKafkaOutput("test", config))
		})
	}
}

func TestDefaultOutputManager_InitOutputs(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "TestDefaultOutputManager_InitOutputs empty config",
			wantErr: false,
		},
		{
			name:    "TestDefaultOutputManager_InitOutputs arango output",
			wantErr: true,
		},
		{
			name:    "TestDefaultOutputManager_InitOutputs kafka output",
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
			manager := NewDefaultOutputManager(config)
			if tt.wantErr {
				assert.Error(t, manager.InitOutputs())
			} else {
				assert.NoError(t, manager.InitOutputs())
			}
		})
	}
}

func TestDefaultOutputManager_GetOutputResources(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "TestDefaultOutputManager_GetOutputResources success",
			wantErr: false,
		},
		{
			name:    "TestDefaultOutputManager_GetOutputResources error",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := config.NewDefaultConfig("../../tests/config/valid-config.yaml")
			manager := NewDefaultOutputManager(config)
			if tt.wantErr {
				output, err := manager.GetOutputResource("test")
				assert.Nil(t, output)
				assert.Error(t, err)
				return
			}
			manager.outputResources["test"] = NewOutputResource()
			output, err := manager.GetOutputResource("test")
			assert.NotNil(t, output)
			assert.NoError(t, err)
		})
	}
}

func TestDefaultOutputManager_StartOutputs(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "TestDefaultOutputManager_StartOutputs",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := config.NewDefaultConfig("../../tests/config/valid-config.yaml")
			manager := NewDefaultOutputManager(config)
			output := NewMockOutput(gomock.NewController(t))
			output.EXPECT().Start().Return()
			outputResource := NewOutputResource()
			outputResource.Output = output
			manager.outputResources["test"] = outputResource
			manager.StartOutputs()
			manager.wg.Wait()
		})
	}
}

func TestDefaultOutputManager_StopOutputs(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "TestDefaultOutputManager_StopOutputs success",
			wantErr: false,
		},
		{
			name:    "TestDefaultOutputManager_StopOutputs",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := config.NewDefaultConfig("../../tests/config/valid-config.yaml")
			manager := NewDefaultOutputManager(config)
			output := NewMockOutput(gomock.NewController(t))
			if tt.wantErr {
				output.EXPECT().Stop().Return(errors.New("error"))
			} else {
				output.EXPECT().Stop().Return(nil)
			}
			outputResource := NewOutputResource()
			outputResource.Output = output
			manager.outputResources["test"] = outputResource
			manager.StopOutputs()
		})
	}
}
