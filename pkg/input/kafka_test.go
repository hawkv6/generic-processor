package input

import (
	"sync"
	"testing"

	"github.com/IBM/sarama"
	"github.com/hawkv6/generic-processor/pkg/config"
	"github.com/hawkv6/generic-processor/pkg/message"
	"github.com/stretchr/testify/assert"
)

func getValidConsumerMessage() *sarama.ConsumerMessage {
	return &sarama.ConsumerMessage{
		Headers: nil,
		Key:     nil,
		Value: []byte(`
								{
  								"fields": {
    								"delete": false,
    								"state/ip": "2001:db8::1",
    								"state/origin": "static",
    								"state/prefix_length": 64,
    								"state/status": "preferred"
  								},
  								"name": "ipv6-addresses",
  								"tags": {
    								"host": "telegraf",
    								"index": "1",
    								"ip": "2001:db8::1",
    								"name": "GigabitEthernet0/0/0/1",
    								"node": "0/RP0/CPU0",
    								"path": "Cisco-IOS-XR-ipv6-io-oper:ipv6-network/nodes/node/interface-data/vrfs/vrf/ipv6-network-interfaces/ipv6-network-interface",
    								"source": "XR-1",
    								"subscription": "hawkv6-openconfig"
  								},
  								"timestamp": 1704728135
								}
				`),
		Topic: "test",
	}
}

func TestNewKafkaInput(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "TestNewKafkaInput",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := config.KafkaInputConfig{
				Broker: "localhost:9092",
				Topic:  "test",
			}
			assert.NotNil(t, NewKafkaInput(config, make(chan message.Command), make(chan message.Result)), nil)
		})
	}
}

func TestKafkaInput_createConfig(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "TestKafkaInput_createConfig",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := config.KafkaInputConfig{
				Broker: "localhost:9092",
				Topic:  "test",
			}
			input := NewKafkaInput(config, make(chan message.Command), make(chan message.Result))
			input.createConfig()
			assert.NotNil(t, input.saramaConfig)
		})
	}
}

func TestKafkaInput_createConsumer(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "TestKafkaInput_createConsumer with connection error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := config.KafkaInputConfig{
				Broker: "localhost:9092",
				Topic:  "test",
			}
			input := NewKafkaInput(config, make(chan message.Command), make(chan message.Result))
			assert.Error(t, input.createConsumer())
		})
	}
}

func TestKafkaInput_createParitionConsumer(t *testing.T) {
	tests := []struct {
		name    string
		topic   string
		wantErr bool
	}{
		{
			name:    "TestKafkaInput_createParitionConsumer success",
			topic:   "test",
			wantErr: false,
		},
		{
			name:    "TestKafkaInput_createParitionConsumer with error",
			topic:   "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := config.KafkaInputConfig{
				Broker: "localhost:9092",
				Topic:  tt.topic,
			}
			input := NewKafkaInput(config, make(chan message.Command), make(chan message.Result))
			input.saramaConsumer = NewKafkaConsumerMock()
			if tt.wantErr {
				assert.Error(t, input.createParitionConsumer())
				return
			}
			assert.NoError(t, input.createParitionConsumer())
		})
	}
}

func TestKafkaInput_Init(t *testing.T) {
	tests := []struct {
		name  string
		topic string
	}{
		{
			name:  "TestKafkaInput_Init connection error",
			topic: "test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := config.KafkaInputConfig{
				Broker: "localhost:9092",
				Topic:  tt.topic,
			}
			input := NewKafkaInput(config, make(chan message.Command), make(chan message.Result))
			assert.Error(t, input.Init())
		})
	}
}

func TestKafkaInput_UnmarshalTelemetryMessage(t *testing.T) {
	tests := []struct {
		name      string
		msg       *sarama.ConsumerMessage
		wantError bool
	}{
		{
			name:      "TestKafkaInput_UnmarshalTelemetryMessage success",
			msg:       getValidConsumerMessage(),
			wantError: false,
		},
		{
			name: "TestKafkaInput_UnmarshalTelemetryMessage invalid JSON",
			msg: &sarama.ConsumerMessage{
				Headers: nil,
				Key:     nil,
				Value:   []byte("no valid json"),
				Topic:   "test",
			},
			wantError: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := config.KafkaInputConfig{
				Broker: "localhost:9092",
				Topic:  "test",
			}
			input := NewKafkaInput(config, make(chan message.Command), make(chan message.Result))
			telemetryMess, err := input.UnmarshalTelemetryMessage(tt.msg)
			if tt.wantError {
				assert.Nil(t, telemetryMess)
				assert.Error(t, err)
				return
			} else {
				assert.NotNil(t, telemetryMess)
				assert.NoError(t, err)
			}
		})
	}
}

func TestKafkaInput_decodeIpv6Message(t *testing.T) {
	tests := []struct {
		name    string
		msg     *message.TelemetryMessage
		wantErr bool
	}{
		{
			name: "TestKafkaInput_decodeIpv6Message success",
			msg: &message.TelemetryMessage{
				Fields: map[string]interface{}{
					"delete":              false,
					"state/ip":            "2001:db8::1",
					"state/origin":        "static",
					"state/prefix_length": 64,
					"state/status":        "preferred",
				},
				Name: "ipv6-addresses",
				Tags: message.MessageTags{
					Host:          "telegraf",
					Index:         "1",
					IP:            "2001:db8::1",
					InterfaceName: "GigabitEthernet0/0/0/1",
					Node:          "0/RP0/CPU0",
					Path:          "Cisco-IOS-XR-ipv6-io-oper:ipv6-network/nodes/node/interface-data/vrfs/vrf/ipv6-network-interfaces/ipv6-network-interface",
					Source:        "XR-1",
					Subscription:  "hawkv6-openconfig",
				},
			},
			wantErr: false,
		},
		{
			name: "TestKafkaInput_decodeIpv6Message decode error",
			msg: &message.TelemetryMessage{
				Fields: map[string]interface{}{
					"delete":   "false",
					"state/ip": "",
				},
			},
			wantErr: true,
		},
		{
			name: "TestKafkaInput_decodeIpv6Message invalid message",
			msg: &message.TelemetryMessage{
				Fields: map[string]interface{}{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := config.KafkaInputConfig{
				Broker: "localhost:9092",
				Topic:  "test",
			}
			input := NewKafkaInput(config, make(chan message.Command), make(chan message.Result))
			result, err := input.decodeIpv6Message(tt.msg)
			if tt.wantErr {
				assert.Nil(t, result)
				assert.Error(t, err)
				return
			}
			assert.NotNil(t, result)
			assert.NoError(t, err)
		})
	}
}

func TestKafkaInput_decodeInterfaceStatusMessage(t *testing.T) {
	tests := []struct {
		name    string
		msg     *message.TelemetryMessage
		wantErr bool
	}{
		{
			name: "TestKafkaInput_decodeInterfaceStatusMessage success",
			msg: &message.TelemetryMessage{
				Fields: map[string]interface{}{
					"admin_status":                 "up",
					"counters/carrier_transitions": 0,
					"cpu":                          "0",
					"delete":                       false,
					"description":                  "GigabitEthernet0/0/0/1",
					"enabled":                      "true",
					"ifindex":                      1,
					"last_change":                  1704728135,
					"logical":                      "false",
					"loopback_mode":                "false",
					"management":                   "false",
					"mtu":                          1500,
					"name":                         "GigabitEthernet0/0/0/1",
					"oper_status":                  "up",
					"type":                         "ethernetCsmacd",
				},
				Name: "oper-state",
				Tags: message.MessageTags{
					Host:          "telegraf",
					Index:         "1",
					IP:            "2001:db8::1",
					InterfaceName: "GigabitEthernet0/0/0/1",
					Node:          "0/RP0/CPU0",
					Path:          "Cisco-IOS-XR-ipv6-io-oper:ipv6-network/nodes/node/interface-data/vrfs/vrf/ipv6-network-interfaces/ipv6-network-interface",
					Source:        "XR-1",
					Subscription:  "hawkv6-openconfig",
				},
			},
			wantErr: false,
		},
		{
			name: "TestKafkaInput_decodeInterfaceStatusMessage decode error",
			msg: &message.TelemetryMessage{
				Fields: map[string]interface{}{
					"admin_status": true,
				},
			},
			wantErr: true,
		},
		{
			name: "TestKafkaInput_decodeInterfaceStatusMessage invalid message",
			msg: &message.TelemetryMessage{
				Fields: map[string]interface{}{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := config.KafkaInputConfig{
				Broker: "localhost:9092",
				Topic:  "test",
			}
			input := NewKafkaInput(config, make(chan message.Command), make(chan message.Result))
			result, err := input.decodeInterfaceStatusMessage(tt.msg)
			if tt.wantErr {
				assert.Nil(t, result)
				assert.Error(t, err)
				return
			}
			assert.NotNil(t, result)
			assert.NoError(t, err)
		})
	}
}

func TestKafkaInput_processMessage(t *testing.T) {
	tests := []struct {
		name    string
		msg     *sarama.ConsumerMessage
		wantErr bool
	}{
		{
			name:    "TestKafkaInput_processMessage success",
			msg:     getValidConsumerMessage(),
			wantErr: false,
		},
		{
			name: "TestKafkaInput_processMessage invalid oper-state message",
			msg: &sarama.ConsumerMessage{
				Headers: nil,
				Key:     nil,
				Value: []byte(`
								{
  								"fields": {
    								"delete": false,
    								"state/ip": "2001:db8::1",
    								"state/origin": "static",
    								"state/prefix_length": 64,
    								"state/status": "preferred"
  								},
  								"name": "oper-state",
  								"tags": {
    								"host": "telegraf",
    								"index": "1",
    								"ip": "2001:db8::1",
    								"name": "GigabitEthernet0/0/0/1",
    								"node": "0/RP0/CPU0",
    								"path": "Cisco-IOS-XR-ipv6-io-oper:ipv6-network/nodes/node/interface-data/vrfs/vrf/ipv6-network-interfaces/ipv6-network-interface",
    								"source": "XR-1",
    								"subscription": "hawkv6-openconfig"
  								},
  								"timestamp": 1704728135
								}
				`),
				Topic: "test",
			},
			wantErr: true,
		},
		{
			name: "TestKafkaInput_processMessage invalid JSON",
			msg: &sarama.ConsumerMessage{
				Headers: nil,
				Key:     nil,
				Value:   []byte("no valid json"),
				Topic:   "test",
			},
			wantErr: true,
		},
		{
			name: "TestKafkaInput_processMessage unknown message",
			msg: &sarama.ConsumerMessage{
				Headers: nil,
				Key:     nil,
				Value: []byte(`
								{
  								"fields": {
    								"delete": "false", 
    								"ip": "2001:db8::1" 
  								},
  								"name": "wrong-name",
  								"tags": {
    								"host": "telegraf",
    								"index": "1",
    								"ip": "2001:db8::1",
    								"name": "GigabitEthernet0/0/0/1",
    								"node": "0/RP0/CPU0",
    								"path": "Cisco-IOS-XR-ipv6-io-oper:ipv6-network/nodes/node/interface-data/vrfs/vrf/ipv6-network-interfaces/ipv6-network-interface",
    								"source": "XR-1",
    								"subscription": "hawkv6-openconfig"
  								},
  								"timestamp": 1704728135
								}
				`),
				Topic: "test",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := config.KafkaInputConfig{
				Broker: "localhost:9092",
				Topic:  "test",
			}
			input := NewKafkaInput(config, make(chan message.Command), make(chan message.Result))
			if tt.wantErr {
				input.processMessage(tt.msg)
				return
			}
			wg := sync.WaitGroup{}
			wg.Add(1)
			go func() {
				<-input.resultChan
				t.Log("Received result")
				wg.Done()
			}()
			input.processMessage(tt.msg)
			wg.Wait()
		})
	}
}

func TestKafkaInput_StartListening(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "TestKafkaInput_StartListening",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := config.KafkaInputConfig{
				Broker: "localhost:9092",
				Topic:  "test",
			}
			input := NewKafkaInput(config, make(chan message.Command), make(chan message.Result))
			partitionConsumer := NewKafkaParitionConsumerMock()
			input.saramaPartitionConsumer = partitionConsumer
			wg := sync.WaitGroup{}
			wg.Add(1)
			go func() {
				input.StartListening()
				wg.Done()
			}()
			partitionConsumer.messageChan <- getValidConsumerMessage()
			close(partitionConsumer.messageChan)
			wg.Done()
		})
	}
}

func TestKafkaInput_Start(t *testing.T) {
	tests := []struct {
		name    string
		command message.Command
	}{
		{
			name:    "TestKafkaInput_Start start with valid command",
			command: message.KafkaListeningCommand{},
		},
		{
			name:    "TestKafkaInput_Start no start invalid command",
			command: message.BaseCommand{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := config.KafkaInputConfig{
				Broker: "localhost:9092",
				Topic:  "test",
			}
			input := NewKafkaInput(config, make(chan message.Command), make(chan message.Result))
			partitionConsumer := NewKafkaParitionConsumerMock()
			input.saramaPartitionConsumer = partitionConsumer
			wg := sync.WaitGroup{}
			wg.Add(1)
			go func() {
				input.Start()
				wg.Done()
			}()
			input.commandChan <- tt.command
			close(input.quitChan)
			wg.Wait()
		})
	}
}

func TestKafkaInput_Stop(t *testing.T) {
	tests := []struct {
		name                 string
		partitionConsumerErr bool
		consumerErr          bool
	}{
		{
			name:                 "TestKafkaInput_Stop no error",
			partitionConsumerErr: false,
			consumerErr:          false,
		},
		{
			name:                 "TestKafkaInput_Stop partition consumer error",
			partitionConsumerErr: true,
			consumerErr:          false,
		},
		{
			name:                 "TestKafkaInput_Stop consumer error",
			partitionConsumerErr: false,
			consumerErr:          true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := config.KafkaInputConfig{
				Broker: "localhost:9092",
				Topic:  "test",
			}
			input := NewKafkaInput(config, make(chan message.Command), make(chan message.Result))
			kafkaConsumer := NewKafkaConsumerMock()
			kafkaPartitionConsumer := NewKafkaParitionConsumerMock()
			input.saramaConsumer = kafkaConsumer
			input.saramaPartitionConsumer = kafkaPartitionConsumer
			if tt.partitionConsumerErr {
				kafkaPartitionConsumer.ReturnError = true
			}
			if tt.consumerErr {
				kafkaConsumer.returnError = true
			}
			if tt.consumerErr || tt.partitionConsumerErr {
				assert.Error(t, input.Stop())
				return
			}
			wg := sync.WaitGroup{}
			wg.Add(1)
			go func() {
				input.StartListening()
				wg.Done()
			}()
			assert.NoError(t, input.Stop())
			wg.Wait()
		})
	}
}
