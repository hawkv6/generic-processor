package processor

import (
	"errors"
	"testing"

	"github.com/hawkv6/generic-processor/pkg/config"
	"github.com/hawkv6/generic-processor/pkg/input"
	"github.com/hawkv6/generic-processor/pkg/output"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNewDefaultProcessorManager(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "TestNewDefaultProcessorManager",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configLocation := "../../tests/config/valid-config.yaml"
			config := config.NewDefaultConfig(configLocation)
			inputManager := input.NewDefaultInputManager(config)
			outputManager := output.NewDefaultOutputManager(config)
			assert.NotNil(t, NewDefaultProcessorManager(config, inputManager, outputManager))
		})
	}
}

func TestDefaultProcessorManager_getInputResources(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "TestDefaultProcessorManager_getInputResources success",
			wantErr: false,
		},
		{
			name:    "TestDefaultProcessorManager_getInputResources error input not found",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configLocation := "../../tests/config/valid-config.yaml"
			config := config.NewDefaultConfig(configLocation)
			err := config.Read()
			assert.Nil(t, err)
			err = config.Validate()
			assert.Nil(t, err)
			inputManager := input.NewMockInputManager(gomock.NewController(t))
			if tt.wantErr {
				inputManager.EXPECT().GetInputResources(gomock.Any()).Return(nil, errors.New("input not found")).AnyTimes()
			} else {
				inputManager.EXPECT().GetInputResources(gomock.Any()).Return(&input.InputResource{}, nil).AnyTimes()
			}
			manager := NewDefaultProcessorManager(config, inputManager, nil)
			inputResources, err := manager.getInputResources()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, inputResources)
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, inputResources)
			}
		})
	}
}

func TestDefaultProcessorManager_getOutputResources(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "TestDefaultProcessorManager_getOutputResources success",
			wantErr: false,
		},
		{
			name:    "TestDefaultProcessorManager_getOutputResources error output not found",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configLocation := "../../tests/config/valid-config.yaml"
			config := config.NewDefaultConfig(configLocation)
			err := config.Read()
			assert.Nil(t, err)
			err = config.Validate()
			assert.Nil(t, err)
			outputManager := output.NewMockOutputManager(gomock.NewController(t))
			if tt.wantErr {
				outputManager.EXPECT().GetOutputResource(gomock.Any()).Return(nil, errors.New("output not found")).AnyTimes()
			} else {
				outputManager.EXPECT().GetOutputResource(gomock.Any()).Return(&output.OutputResource{}, nil).AnyTimes()
			}
			manager := NewDefaultProcessorManager(config, nil, outputManager)
			outputResources, err := manager.getOutputResources()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, outputResources)
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, outputResources)
			}
		})
	}
}

func TestDefaultProcessorManager_Init(t *testing.T) {
	tests := []struct {
		name          string
		wantInputErr  bool
		wantOutputErr bool
	}{
		{
			name:          "TestDefaultProcessorManager_Init success",
			wantInputErr:  false,
			wantOutputErr: false,
		},
		{
			name:          "TestDefaultProcessorManager_Init error input not found",
			wantInputErr:  true,
			wantOutputErr: false,
		},
		{
			name:          "TestDefaultProcessorManager_Init error output not found",
			wantInputErr:  false,
			wantOutputErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configLocation := "../../tests/config/valid-config.yaml"
			config := config.NewDefaultConfig(configLocation)
			err := config.Read()
			assert.Nil(t, err)
			err = config.Validate()
			assert.Nil(t, err)
			inputManager := input.NewMockInputManager(gomock.NewController(t))
			if tt.wantInputErr {
				inputManager.EXPECT().GetInputResources(gomock.Any()).Return(nil, errors.New("input not found")).AnyTimes()
			} else {
				inputManager.EXPECT().GetInputResources(gomock.Any()).Return(&input.InputResource{}, nil).AnyTimes()
			}
			outputManager := output.NewMockOutputManager(gomock.NewController(t))
			if tt.wantOutputErr {
				outputManager.EXPECT().GetOutputResource(gomock.Any()).Return(nil, errors.New("output not found")).AnyTimes()
			} else {
				outputManager.EXPECT().GetOutputResource(gomock.Any()).Return(&output.OutputResource{}, nil).AnyTimes()
			}
			manager := NewDefaultProcessorManager(config, inputManager, outputManager)
			err = manager.Init()
			if tt.wantInputErr || tt.wantOutputErr {
				assert.Error(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestDefaultProcessorManager_StartProcesssors(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "TestDefaultProcessorManager_StartProcessors",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configLocation := "../../tests/config/valid-config.yaml"
			config := config.NewDefaultConfig(configLocation)
			err := config.Read()
			assert.Nil(t, err)
			err = config.Validate()
			assert.Nil(t, err)
			inputManager := input.NewDefaultInputManager(config)
			outputManager := output.NewDefaultOutputManager(config)
			manager := NewDefaultProcessorManager(config, inputManager, outputManager)
			processorMock := NewMockProcessor(gomock.NewController(t))
			processorMock.EXPECT().Start().Return().AnyTimes()
			processorMock.EXPECT().Stop().Return().AnyTimes()
			manager.processors = map[string]Processor{
				"test": processorMock,
			}
			manager.StartProcessors()
			for _, processor := range manager.processors {
				processor.Stop()
			}
		})
	}
}

func TestDefaultManager_StopProcessors(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "TestDefaultManager_StopProcessors",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configLocation := "../../tests/config/valid-config.yaml"
			config := config.NewDefaultConfig(configLocation)
			err := config.Read()
			assert.Nil(t, err)
			err = config.Validate()
			assert.Nil(t, err)
			inputManager := input.NewDefaultInputManager(config)
			outputManager := output.NewDefaultOutputManager(config)
			manager := NewDefaultProcessorManager(config, inputManager, outputManager)
			processorMock := NewMockProcessor(gomock.NewController(t))
			processorMock.EXPECT().Start().Return().AnyTimes()
			processorMock.EXPECT().Stop().Return().AnyTimes()
			manager.processors = map[string]Processor{
				"test": processorMock,
			}
			manager.StartProcessors()
			manager.StopProcessors()
		})
	}
}
