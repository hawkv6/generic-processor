package processor

import (
	"sync"
	"testing"
	"time"

	"github.com/hawkv6/generic-processor/pkg/config"
	"github.com/hawkv6/generic-processor/pkg/input"
	"github.com/hawkv6/generic-processor/pkg/message"
	"github.com/hawkv6/generic-processor/pkg/output"
	"github.com/stretchr/testify/assert"
)

func TestNewTelemetryToArangoProcessor(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "Test case 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotNil(t, NewTelemetryToArangoProcessor(config.TelemetryToArangoProcessorConfig{}, make(map[string]input.InputResource), make(map[string]output.OutputResource), NewKafkaOpenConfigProcessor(), NewMinMaxNormalizer()))
		})
	}
}

func TestTelemetryToArangoProcessor_sendInfluxCommands(t *testing.T) {
	config := config.TelemetryToArangoProcessorConfig{
		BaseProcessorConfig: config.BaseProcessorConfig{
			Name: "Test",
		},
		Inputs:   []string{"input"},
		Outputs:  []string{"output"},
		Interval: 1,
		Normalization: config.Normalization{
			FieldMappings: map[string]string{
				"field1": "field2",
			},
		},
		Modes: []config.Mode{
			{
				InputOptions: map[string]config.InputOption{
					"input": {
						Name:        "input",
						Measurement: "measurement",
						Field:       "field",
						Method:      "method",
						GroupBy:     []string{"group"},
						Transformation: &config.Transformation{
							Operation: "operation",
							Period:    1,
						},
					},
				},
				OutputOptions: map[string]config.OutputOption{
					"output": {
						Name:       "output",
						Method:     "method",
						Collection: "collection",
						FilterBy:   []string{"filter"},
						Field:      "field",
					},
				},
			},
		},
	}

	tests := []struct {
		name string
	}{
		{
			name: "TestTelemetryToArangoProcessor",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := NewTelemetryToArangoProcessor(config, make(map[string]input.InputResource), make(map[string]output.OutputResource), NewKafkaOpenConfigProcessor(), NewMinMaxNormalizer())
			commandChan := make(chan message.Command)
			wg := sync.WaitGroup{}
			wg.Add(1)
			go func() {
				command := <-commandChan
				influxQueryCommand := command.(message.InfluxQueryCommand)
				assert.Equal(t, config.Modes[0].InputOptions["input"].Measurement, influxQueryCommand.Measurement)
				assert.Equal(t, config.Modes[0].InputOptions["input"].Field, influxQueryCommand.Field)
				assert.Equal(t, config.Modes[0].InputOptions["input"].Transformation.Operation, influxQueryCommand.Transformation.Operation)
				assert.Equal(t, config.Modes[0].InputOptions["input"].Transformation.Period, influxQueryCommand.Transformation.Period)
				assert.Equal(t, config.Modes[0].InputOptions["input"].Method, influxQueryCommand.Method)
				assert.Equal(t, config.Modes[0].InputOptions["input"].GroupBy, influxQueryCommand.GroupBy)
				assert.Equal(t, config.Interval, influxQueryCommand.Interval)
				assert.Equal(t, config.Modes[0].OutputOptions, influxQueryCommand.OutputOptions)
				wg.Done()
			}()
			processor.sendInfluxCommands("input", commandChan)
			wg.Wait()
		})
	}
}

func TestTelemetryToArangoProcessor_startSchedulingInfluxCommands(t *testing.T) {
	config := config.TelemetryToArangoProcessorConfig{
		BaseProcessorConfig: config.BaseProcessorConfig{
			Name: "Test",
		},
		Inputs:   []string{"input"},
		Outputs:  []string{"output"},
		Interval: 1,
		Normalization: config.Normalization{
			FieldMappings: map[string]string{
				"field1": "field2",
			},
		},
		Modes: []config.Mode{
			{
				InputOptions: map[string]config.InputOption{
					"input": {
						Name:        "input",
						Measurement: "measurement",
						Field:       "field",
						Method:      "method",
						GroupBy:     []string{"group"},
						Transformation: &config.Transformation{
							Operation: "operation",
							Period:    1,
						},
					},
				},
				OutputOptions: map[string]config.OutputOption{
					"output": {
						Name:       "output",
						Method:     "method",
						Collection: "collection",
						FilterBy:   []string{"filter"},
						Field:      "field",
					},
				},
			},
		},
	}

	tests := []struct {
		name string
	}{
		{
			name: "TestTelemetryToArangoProcessor",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := NewTelemetryToArangoProcessor(config, make(map[string]input.InputResource), make(map[string]output.OutputResource), NewKafkaOpenConfigProcessor(), NewMinMaxNormalizer())
			commandChan := make(chan message.Command)
			wg := sync.WaitGroup{}
			wg.Add(1)
			go func() {
				command := <-commandChan
				influxQueryCommand := command.(message.InfluxQueryCommand)
				assert.Equal(t, config.Modes[0].InputOptions["input"].Measurement, influxQueryCommand.Measurement)
				assert.Equal(t, config.Modes[0].InputOptions["input"].Field, influxQueryCommand.Field)
				assert.Equal(t, config.Modes[0].InputOptions["input"].Transformation.Operation, influxQueryCommand.Transformation.Operation)
				assert.Equal(t, config.Modes[0].InputOptions["input"].Transformation.Period, influxQueryCommand.Transformation.Period)
				assert.Equal(t, config.Modes[0].InputOptions["input"].Method, influxQueryCommand.Method)
				assert.Equal(t, config.Modes[0].InputOptions["input"].GroupBy, influxQueryCommand.GroupBy)
				assert.Equal(t, config.Interval, influxQueryCommand.Interval)
				assert.Equal(t, config.Modes[0].OutputOptions, influxQueryCommand.OutputOptions)
			}()
			go func() {
				processor.startSchedulingInfluxCommands(config.Inputs[0], commandChan)
				wg.Done()
			}()
			time.Sleep(1 * time.Second)
			close(processor.quitChan)
			wg.Wait()
		})
	}
}

func TestTelemetryToArangoProcessor_sendArangoUpdateCommands(t *testing.T) {
	config := config.TelemetryToArangoProcessorConfig{
		BaseProcessorConfig: config.BaseProcessorConfig{
			Name: "Test",
		},
		Inputs:   []string{"input"},
		Outputs:  []string{"output"},
		Interval: 1,
		Normalization: config.Normalization{
			FieldMappings: map[string]string{
				"field1": "field2",
			},
		},
		Modes: []config.Mode{
			{
				InputOptions: map[string]config.InputOption{
					"input": {
						Name:        "input",
						Measurement: "measurement",
						Field:       "field",
						Method:      "method",
						GroupBy:     []string{"group"},
						Transformation: &config.Transformation{
							Operation: "operation",
							Period:    1,
						},
					},
				},
				OutputOptions: map[string]config.OutputOption{
					"output": {
						Name:       "output",
						Method:     "method",
						Collection: "collection",
						FilterBy:   []string{"filter"},
						Field:      "field",
					},
				},
			},
		},
	}

	arangoUpdateCommands := make(map[string]map[string]message.ArangoUpdateCommand)
	arangoUpdateCommands["output"] = map[string]message.ArangoUpdateCommand{
		"collection": {
			Collection: "collection",
			StatisticalData: map[string]float64{
				"field": 1.0,
			},
			Updates: map[string]message.ArangoUpdate{
				"field": {
					Tags: map[string]string{
						"tag": "tag",
					},
					Fields: []string{"field"},
					Values: []float64{1.0},
				},
			},
		},
	}

	tests := []struct {
		name string
	}{
		{
			name: "TestTelemetryToArangoProcessor",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := NewTelemetryToArangoProcessor(config, make(map[string]input.InputResource), make(map[string]output.OutputResource), NewKafkaOpenConfigProcessor(), NewMinMaxNormalizer())
			outputResource := output.NewOutputResource()
			processor.outputResources["output"] = outputResource
			wg := sync.WaitGroup{}
			wg.Add(1)
			go func() {
				command := <-outputResource.CommandChan
				arangoUpdateCommand := command.(message.ArangoUpdateCommand)
				assert.Equal(t, arangoUpdateCommands["output"]["collection"].Collection, arangoUpdateCommand.Collection)
				assert.Equal(t, arangoUpdateCommands["output"]["collection"].StatisticalData, arangoUpdateCommand.StatisticalData)
				assert.Equal(t, arangoUpdateCommands["output"]["collection"].Updates, arangoUpdateCommand.Updates)
				wg.Done()
			}()
			processor.sendArangoUpdateCommands(arangoUpdateCommands)
			wg.Wait()
		})
	}
}

func TestTelemetryToArangoProcessor_checkValidInterfaceName(t *testing.T) {
	tests := []struct {
		name    string
		isValid bool
		tags    map[string]string
	}{
		{
			name: "TestTelemetryToArangoProcessor valid",
			tags: map[string]string{
				"tag1":           "tag1",
				"interface_name": "GigabitEthernet0/0/0/0",
			},
			isValid: true,
		},
		{
			name: "TestTelemetryToArangoProcessor invalid",
			tags: map[string]string{
				"tag1": "tag1",
				"tag2": "tag2",
			},
			isValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := NewTelemetryToArangoProcessor(config.TelemetryToArangoProcessorConfig{}, make(map[string]input.InputResource), make(map[string]output.OutputResource), NewKafkaOpenConfigProcessor(), NewMinMaxNormalizer())
			assert.NotNil(t, processor)
			assert.Equal(t, tt.isValid, processor.checkValidInterfaceName(tt.tags))
		})
	}
}

func TestTelemetryToArangoProcessor_generateKeyFromFilter(t *testing.T) {
	testIp := "2001:db8::1"
	outputOption := config.OutputOption{
		Name:       "output",
		Method:     "update",
		Collection: "ls_link",
		FilterBy:   []string{"another_field", "local_link_ip"},
		Field:      "unidir_link_delay",
	}

	tests := []struct {
		name        string
		filterBy    []string
		tags        map[string]string
		expectedKey string
	}{
		{
			name:     "TestTelemetryToArangoProcessor valid",
			filterBy: []string{"tag1", "tag2"},
			tags: map[string]string{
				"interface_name": "GigabitEthernet0/0/0/0",
				"source":         "XR-1",
			},
			expectedKey: testIp,
		},
		{
			name:     "TestTelemetryToArangoProcessor missing source tag",
			filterBy: []string{"tag1", "tag2"},
			tags: map[string]string{
				"interface_name": "GigabitEthernet0/0/0/0",
			},
			expectedKey: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kafkaProcessor := NewKafkaOpenConfigProcessor()
			processor := NewTelemetryToArangoProcessor(config.TelemetryToArangoProcessorConfig{}, make(map[string]input.InputResource), make(map[string]output.OutputResource), kafkaProcessor, NewMinMaxNormalizer())
			assert.NotNil(t, processor)
			kafkaProcessor.activeIpv6Addresses["XR-1GigabitEthernet0/0/0/0"] = testIp
			assert.Equal(t, tt.expectedKey, processor.generateKeyFromFilter(outputOption.FilterBy, tt.tags))
		})
	}
}

func TestTelemetryToArangoProcessor_addArangoUpdate(t *testing.T) {

	tests := []struct {
		name    string
		fields  []string
		results []message.InfluxResult
		key     string
		config  config.TelemetryToArangoProcessorConfig
		command *message.ArangoUpdateCommand
	}{
		{
			name:   "TestTelemetryToArangoProcessor new update with normalization data",
			fields: []string{"unidir_link_delay", "unidir_packet_loss"},
			results: []message.InfluxResult{
				{
					Tags: map[string]string{
						"interface_name": "GigabitEthernet0/0/0/0",
						"source":         "XR-1",
					},
					Value: 1.0,
				},
				{
					Tags: map[string]string{
						"interface_name": "GigabitEthernet0/0/0/0",
						"source":         "XR-1",
					},
					Value: 2.0,
				},
			},
			key: "2001:db8::1",
			config: config.TelemetryToArangoProcessorConfig{
				BaseProcessorConfig: config.BaseProcessorConfig{
					Name: "Test",
				},
				Inputs:   []string{"input"},
				Outputs:  []string{"output"},
				Interval: 1,
				Normalization: config.Normalization{
					FieldMappings: map[string]string{
						"unidir_link_delay": "normalized_unidir_link_delay",
					},
				},
			},
			command: message.NewArangoUpdateCommand("ls_link"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := NewTelemetryToArangoProcessor(tt.config, make(map[string]input.InputResource), make(map[string]output.OutputResource), NewKafkaOpenConfigProcessor(), NewMinMaxNormalizer())
			assert.NotNil(t, processor)
			for i, result := range tt.results {
				processor.addArangoUpdate(tt.command.Updates, tt.key, result, tt.fields[i])
			}
			assert.Equal(t, 1, len(tt.command.Updates))
			assert.Equal(t, tt.command.Updates[tt.key].Fields, []string{"unidir_link_delay", "unidir_packet_loss"})
			assert.Equal(t, tt.command.Updates[tt.key].Values, []float64{1.0, 2.0})
		})
	}
}

func TestTeleemtryToArangoProcessor_addStatisticalData(t *testing.T) {
	tests := []struct {
		name    string
		field   string
		config  config.TelemetryToArangoProcessorConfig
		command message.ArangoUpdateCommand
	}{
		{
			name:  "TestTelemetryToArangoProcessor new update with normalization data",
			field: "unidir_link_delay",
			config: config.TelemetryToArangoProcessorConfig{
				BaseProcessorConfig: config.BaseProcessorConfig{
					Name: "Test",
				},
				Inputs:   []string{"input"},
				Outputs:  []string{"output"},
				Interval: 1,
				Normalization: config.Normalization{
					FieldMappings: map[string]string{
						"unidir_link_delay": "normalized_unidir_link_delay",
					},
				},
			},
			command: *message.NewArangoUpdateCommand("ls_link"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			minMaxNormalizer := NewMinMaxNormalizer()
			processor := NewTelemetryToArangoProcessor(tt.config, make(map[string]input.InputResource), make(map[string]output.OutputResource), NewKafkaOpenConfigProcessor(), minMaxNormalizer)
			assert.NotNil(t, processor)
			for _, value := range []float64{1.0, 2.0, 3.0, 4.0, 5.0} {
				minMaxNormalizer.AddDataToNormalize("unidir_link_delay", value)
			}
			processor.addStatisticalData(tt.field, tt.command)
			assert.Equal(t, 1.5, tt.command.StatisticalData["unidir_link_delay_q1"])
			assert.Equal(t, 4.5, tt.command.StatisticalData["unidir_link_delay_q3"])
			assert.Equal(t, 3.0, tt.command.StatisticalData["unidir_link_delay_iqr"])
			assert.Equal(t, 1.0, tt.command.StatisticalData["unidir_link_delay_lower_fence"])
			assert.Equal(t, 5.0, tt.command.StatisticalData["unidir_link_delay_upper_fence"])
		})
	}
}

func TestTelemetryToArangoProcessor_createArangoUpdateCommands(t *testing.T) {
	testIp := "2001:db8::1"
	outputOption := config.OutputOption{
		Name:       "output",
		Method:     "update",
		Collection: "ls_link",
		FilterBy:   []string{"another_field", "local_link_ip"},
		Field:      "unidir_link_delay",
	}
	results := []message.InfluxResult{
		{
			Tags: map[string]string{
				"interface_name": "Loopback0",
				"source":         "XR-1",
			},
			Value: 2.0,
		},
		{
			Tags: map[string]string{
				"interface_name": "GigabitEthernet0/0/0/0",
				"source":         "XR-2",
			},
			Value: 1.0,
		},
		{
			Tags: map[string]string{
				"interface_name": "GigabitEthernet0/0/0/0",
				"source":         "XR-1",
			},
			Value: 2.0,
		},
	}

	command := *message.NewArangoUpdateCommand("ls_link")

	tests := []struct {
		name    string
		config  config.TelemetryToArangoProcessorConfig
		results []message.InfluxResult
		command message.ArangoUpdateCommand
	}{
		{
			name: "TestTelemetryToArangoProcessor_createArangoUpdateCommands success",
			config: config.TelemetryToArangoProcessorConfig{
				BaseProcessorConfig: config.BaseProcessorConfig{
					Name: "Test",
				},
				Inputs:   []string{"input"},
				Outputs:  []string{"output"},
				Interval: 1,
				Normalization: config.Normalization{
					FieldMappings: map[string]string{
						"unidir_link_delay": "normalized_unidir_link_delay",
					},
				},
			},
			results: results,
			command: command,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kafkaProcessor := NewKafkaOpenConfigProcessor()
			processor := NewTelemetryToArangoProcessor(tt.config, make(map[string]input.InputResource), make(map[string]output.OutputResource), kafkaProcessor, NewMinMaxNormalizer())
			kafkaProcessor.activeIpv6Addresses["XR-1GigabitEthernet0/0/0/0"] = testIp
			assert.NotNil(t, processor)
			processor.createArangoUpdateCommands(outputOption, tt.results, tt.command)
			assert.Equal(t, 1, len(tt.command.Updates))
			assert.Equal(t, tt.command.Updates["2001:db8::1"].Fields, []string{"unidir_link_delay"})
			assert.Equal(t, tt.command.Updates["2001:db8::1"].Values, []float64{2.0})
		})
	}
}

func TestTelemetryToArangoProcessor_processInfluxResultMessage(t *testing.T) {
	tests := []struct {
		name             string
		noOutputResource bool
	}{
		{
			name:             "TestTelemetryToArangoProcessor_processInfluxResultMessage success",
			noOutputResource: false,
		},
		{
			name:             "TestTelemetryToArangoProcessor_processInfluxResultMessage success",
			noOutputResource: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := NewTelemetryToArangoProcessor(config.TelemetryToArangoProcessorConfig{}, make(map[string]input.InputResource), make(map[string]output.OutputResource), NewKafkaOpenConfigProcessor(), NewMinMaxNormalizer())
			assert.NotNil(t, processor)
			msg := message.InfluxResultMessage{
				OutputOptions: map[string]config.OutputOption{
					"output": {
						Name:       "output",
						Method:     "update",
						Collection: "ls_link",
						FilterBy:   []string{"another_field", "local_link_ip"},
						Field:      "unidir_link_delay",
					},
				},
				Results: []message.InfluxResult{
					{
						Tags: map[string]string{
							"interface_name": "GigabitEthernet0/0/0/0",
							"source":         "XR-1",
						},
						Value: 2.0,
					},
				},
			}
			arangoUpdateCommands := make(map[string]map[string]message.ArangoUpdateCommand)
			if tt.noOutputResource {
				processor.processInfluxResultMessage(msg, arangoUpdateCommands)
				return
			}
			outputResouce := output.NewOutputResource()
			commandChan := make(chan message.Command)
			resultChan := make(chan message.Result)
			outputResouce.Output = output.NewArangoOutput(config.ArangoOutputConfig{}, commandChan, resultChan)
			processor.outputResources["output"] = outputResouce
			processor.processInfluxResultMessage(msg, arangoUpdateCommands)
		})
	}
}

func TestTelemetryToArangoProcessor_processInfluxResultMessages(t *testing.T) {
	tests := []struct {
		name      string
		inputName string
	}{
		{
			name:      "TestTelemetryToArangoProcessor_processInfluxResultMessages success",
			inputName: "input",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := NewTelemetryToArangoProcessor(config.TelemetryToArangoProcessorConfig{}, make(map[string]input.InputResource), make(map[string]output.OutputResource), NewKafkaOpenConfigProcessor(), NewMinMaxNormalizer())
			assert.NotNil(t, processor)
			msg := message.InfluxResultMessage{
				OutputOptions: map[string]config.OutputOption{
					"output": {
						Name:       "output",
						Method:     "update",
						Collection: "ls_link",
						FilterBy:   []string{"another_field", "local_link_ip"},
						Field:      "unidir_link_delay",
					},
				},
				Results: []message.InfluxResult{
					{
						Tags: map[string]string{
							"interface_name": "GigabitEthernet0/0/0/0",
							"source":         "XR-1",
						},
						Value: 2.0,
					},
				},
			}
			msgs := []message.Result{msg, message.BaseResultMessage{}}
			processor.processInfluxResultMessages(tt.inputName, msgs)
		})
	}
}

func TestTelemetryToArangoProcessor_startInfluxProcessing(t *testing.T) {
	tests := []struct {
		name        string
		inputName   string
		commandChan chan message.Command
		resultChan  chan message.Result
	}{
		{
			name:        "TestTelemetryToArangoProcessor_startInfluxProcessing success",
			commandChan: make(chan message.Command),
			resultChan:  make(chan message.Result),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := config.TelemetryToArangoProcessorConfig{
				Modes: []config.Mode{
					{
						InputOptions: map[string]config.InputOption{
							"input": {
								Name:        "input",
								Measurement: "measurement",
								Field:       "field",
								Method:      "method",
								GroupBy:     []string{"group"},
								Transformation: &config.Transformation{
									Operation: "operation",
									Period:    1,
								},
							},
						},
					},
				},
				Interval: 1,
			}
			processor := NewTelemetryToArangoProcessor(config, make(map[string]input.InputResource), make(map[string]output.OutputResource), NewKafkaOpenConfigProcessor(), NewMinMaxNormalizer())
			assert.NotNil(t, processor)
			wg := sync.WaitGroup{}
			wg.Add(1)
			go func() {
				processor.startInfluxProcessing(tt.inputName, tt.commandChan, tt.resultChan)
				wg.Done()
			}()
			tt.resultChan <- message.BaseResultMessage{}
			time.Sleep(100 * time.Millisecond)
			close(processor.quitChan)
			wg.Wait()
		})
	}
}

func TestTelemetryToArangoProcessor_processArangoEventNotificationMessage(t *testing.T) {
	tests := []struct {
		name     string
		eventMsg message.ArangoEventNotificationMessage
	}{
		{
			name: "TestTelemetryToArangoProcessor_processArangoEVentNotificationMessage success",
			eventMsg: message.ArangoEventNotificationMessage{
				EventMessages: []message.EventMessage{
					{
						Key:       "key",
						Id:        "id",
						TopicType: 1,
					},
				},
				Action: "update",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := NewTelemetryToArangoProcessor(config.TelemetryToArangoProcessorConfig{}, make(map[string]input.InputResource), make(map[string]output.OutputResource), NewKafkaOpenConfigProcessor(), NewMinMaxNormalizer())
			assert.NotNil(t, processor)
			wg := sync.WaitGroup{}
			wg.Add(1)
			outputResouce := output.NewOutputResource()
			kafkaOutputs := map[string]output.OutputResource{
				"output": outputResouce,
			}
			outputResouce.Output = output.NewKafkaOutput(config.KafkaOutputConfig{}, outputResouce.CommandChan, nil)
			go func() {
				command := <-outputResouce.CommandChan
				assert.NotNil(t, command)
				kafkaEventCommand := command.(message.KafkaEventCommand)
				assert.Equal(t, tt.eventMsg.EventMessages[0].Id, kafkaEventCommand.Updates[0].Id)
				assert.Equal(t, tt.eventMsg.EventMessages[0].Key, kafkaEventCommand.Updates[0].Key)
				assert.Equal(t, tt.eventMsg.EventMessages[0].TopicType, kafkaEventCommand.Updates[0].TopicType)
				assert.Equal(t, tt.eventMsg.Action, kafkaEventCommand.Updates[0].Action)
				wg.Done()
			}()
			processor.processArangoEventNotificationMessage(tt.eventMsg, kafkaOutputs)
			wg.Wait()
		})
	}
}

func TestTelemetryToArangoProcessor_processArangoNormalizationMessage(t *testing.T) {
	tests := []struct {
		name      string
		normalMsg message.ArangoNormalizationMessage
	}{
		{
			name: "TestTelemetryToArangoProcessor_processArangoNormalizationMessage success",
			normalMsg: message.ArangoNormalizationMessage{
				Measurement: "measurement",
				NormalizationMessages: []message.NormalizationMessage{
					{
						Tags: map[string]string{
							"interface_name": "GigabitEthernet0/0/0/0",
							"source":         "XR-1",
						},
						Fields: map[string]float64{
							"normalized_unidir_link_delay": 1.0,
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := NewTelemetryToArangoProcessor(config.TelemetryToArangoProcessorConfig{}, make(map[string]input.InputResource), make(map[string]output.OutputResource), NewKafkaOpenConfigProcessor(), NewMinMaxNormalizer())
			assert.NotNil(t, processor)
			wg := sync.WaitGroup{}
			wg.Add(1)
			outputResouce := output.NewOutputResource()
			kafkaOutputs := map[string]output.OutputResource{
				"output": outputResouce,
			}
			outputResouce.Output = output.NewKafkaOutput(config.KafkaOutputConfig{}, outputResouce.CommandChan, nil)
			go func() {
				command := <-outputResouce.CommandChan
				assert.NotNil(t, command)
				kafkaNormalizationCommand := command.(message.KafkaNormalizationCommand)
				assert.Equal(t, tt.normalMsg.NormalizationMessages[0].Tags, kafkaNormalizationCommand.Updates[0].Tags)
				assert.Equal(t, tt.normalMsg.NormalizationMessages[0].Fields, kafkaNormalizationCommand.Updates[0].Fields)
				wg.Done()
			}()
			processor.processArangoNormalizationMessage(tt.normalMsg, kafkaOutputs)
			wg.Wait()
		})
	}
}

func TestTelemetryToArangoProcessor_processArangoResultMessages(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "TestTelemetryToArangoProcessor_processArangoResultMessages",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := NewTelemetryToArangoProcessor(config.TelemetryToArangoProcessorConfig{}, make(map[string]input.InputResource), make(map[string]output.OutputResource), NewKafkaOpenConfigProcessor(), NewMinMaxNormalizer())
			assert.NotNil(t, processor)
			outputResouce := output.NewOutputResource()
			kafkaOutputs := map[string]output.OutputResource{
				"output": outputResouce,
			}
			outputResouce.Output = output.NewKafkaOutput(config.KafkaOutputConfig{}, outputResouce.CommandChan, nil)
			resultChan := make(chan message.Result)
			arangoResultChannels := map[string]chan message.Result{
				"arango": resultChan,
			}
			wg := sync.WaitGroup{}
			wg.Add(2)
			go func() {
				processor.processArangoResultMessages(kafkaOutputs, arangoResultChannels)
				wg.Done()
			}()
			go func() {
				command := <-outputResouce.CommandChan
				assert.NotNil(t, command)
				command = <-outputResouce.CommandChan
				assert.NotNil(t, command)
				wg.Done()
			}()
			resultChan <- message.ArangoEventNotificationMessage{}
			resultChan <- message.ArangoNormalizationMessage{}
			resultChan <- message.BaseResultMessage{}
			time.Sleep(100 * time.Millisecond)
			close(processor.quitChan)
			wg.Wait()
		})
	}
}

func TestTelemetryToArangoProcessor_Start(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "TestTelemetryToArangoProcessor_Start",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kafkaInputResources := input.NewInputResource()
			kafkaInputResources.Input = input.NewKafkaInput(config.KafkaInputConfig{}, kafkaInputResources.CommandChan, kafkaInputResources.ResultChan)
			influxInputResources := input.NewInputResource()
			influxInput, _ := input.NewInfluxInput(config.InfluxInputConfig{}, influxInputResources.CommandChan, influxInputResources.ResultChan)
			influxInputResources.Input = influxInput
			inputResources := map[string]input.InputResource{
				"kafka":  kafkaInputResources,
				"influx": influxInputResources,
			}
			arangoOutputResource := output.NewOutputResource()
			arangoOutputResource.Output = output.NewArangoOutput(config.ArangoOutputConfig{}, arangoOutputResource.CommandChan, arangoOutputResource.ResultChan)
			kafkaOutputResource := output.NewOutputResource()
			kafkaOutputResource.Output = output.NewKafkaOutput(config.KafkaOutputConfig{}, kafkaOutputResource.CommandChan, kafkaOutputResource.ResultChan)
			outputResources := map[string]output.OutputResource{
				"arango": arangoOutputResource,
				"kafka":  kafkaOutputResource,
			}

			processor := NewTelemetryToArangoProcessor(config.TelemetryToArangoProcessorConfig{Interval: 1000}, inputResources, outputResources, NewKafkaOpenConfigProcessor(), NewMinMaxNormalizer())
			assert.NotNil(t, processor)
			wg := sync.WaitGroup{}
			wg.Add(1)
			go func() {
				processor.Start()
				wg.Done()
			}()
			time.Sleep(100 * time.Millisecond)
			close(processor.quitChan)
			wg.Wait()
		})
	}
}

func TestTelemetryToArangoProcessor_Stop(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "TestTelemetryToArangoProcessor_Stop",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kafkaInputResources := input.NewInputResource()
			kafkaInputResources.Input = input.NewKafkaInput(config.KafkaInputConfig{}, kafkaInputResources.CommandChan, kafkaInputResources.ResultChan)
			influxInputResources := input.NewInputResource()
			influxInput, _ := input.NewInfluxInput(config.InfluxInputConfig{}, influxInputResources.CommandChan, influxInputResources.ResultChan)
			influxInputResources.Input = influxInput
			inputResources := map[string]input.InputResource{
				"kafka":  kafkaInputResources,
				"influx": influxInputResources,
			}
			arangoOutputResource := output.NewOutputResource()
			arangoOutputResource.Output = output.NewArangoOutput(config.ArangoOutputConfig{}, arangoOutputResource.CommandChan, arangoOutputResource.ResultChan)
			kafkaOutputResource := output.NewOutputResource()
			kafkaOutputResource.Output = output.NewKafkaOutput(config.KafkaOutputConfig{}, kafkaOutputResource.CommandChan, kafkaOutputResource.ResultChan)
			outputResources := map[string]output.OutputResource{
				"arango": arangoOutputResource,
				"kafka":  kafkaOutputResource,
			}

			processor := NewTelemetryToArangoProcessor(config.TelemetryToArangoProcessorConfig{Interval: 1000}, inputResources, outputResources, NewKafkaOpenConfigProcessor(), NewMinMaxNormalizer())
			assert.NotNil(t, processor)
			wg := sync.WaitGroup{}
			wg.Add(1)
			go func() {
				processor.Start()
				wg.Done()
			}()
			time.Sleep(100 * time.Millisecond)
			processor.Stop()
			wg.Wait()
		})
	}
}
