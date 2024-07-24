package output

import (
	"sync"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/hawkv6/generic-processor/pkg/config"
	"github.com/hawkv6/generic-processor/pkg/message"
	"github.com/stretchr/testify/assert"
)

func TestNewKafkaOutput(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "TestNewKafkaOutput",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := config.KafkaOutputConfig{
				Broker: "localhost:9092",
				Topic:  "test",
			}
			assert.NotNil(t, NewKafkaOutput(config, make(chan message.Command), make(chan message.Result)))
		})
	}
}
func TestKafkaOutput_createConfig(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "TestKafkaOutput_createConfig",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := config.KafkaOutputConfig{
				Broker: "localhost:9092",
				Topic:  "test",
			}
			input := NewKafkaOutput(config, make(chan message.Command), make(chan message.Result))
			input.createConfig()
			assert.NotNil(t, input.kafkaConfig)
		})
	}
}

func TestKafkaOutput_Init(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "TestKafkaOutput_Init",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := config.KafkaOutputConfig{
				Broker: "localhost:9092",
				Topic:  "test",
			}
			output := NewKafkaOutput(config, make(chan message.Command), make(chan message.Result))
			err := output.Init()
			assert.NotNil(t, err)
		})
	}
}

func TestKafkaOutput_publishMessage(t *testing.T) {
	tests := []struct {
		name           string
		msg            message.KafkaEventMessage
		wantMarshalErr bool
	}{
		{
			name: "TestKafkaOutput_publishMessage",
			msg: message.KafkaEventMessage{
				TopicType: 1,
				Key:       "key",
				Id:        "id",
				Action:    "action",
			},
			wantMarshalErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := config.KafkaOutputConfig{
				Broker: "localhost:9092",
				Topic:  "test",
			}
			output := NewKafkaOutput(config, make(chan message.Command), make(chan message.Result))
			producer := NewKafkaProducerMock()
			output.producer = producer
			wg := sync.WaitGroup{}
			wg.Add(1)
			go func() {
				<-producer.msgChan
				wg.Done()
			}()
			output.publishMessage(tt.msg)
			wg.Wait()
		})
	}
}

func TestKafkaOutput_sortTags(t *testing.T) {
	tests := []struct {
		name string
		msg  message.KafkaNormalizationMessage
	}{
		{
			name: "TestKafkaOutput_sortTags",
			msg: message.KafkaNormalizationMessage{
				Measurement: "measurement",
				Tags: map[string]string{
					"C": "value3",
					"A": "value1",
					"B": "value2",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := config.KafkaOutputConfig{
				Broker: "localhost:9092",
				Topic:  "test",
			}
			output := NewKafkaOutput(config, make(chan message.Command), make(chan message.Result))
			assert.NotNil(t, output.sortTags(tt.msg))
			assert.Equal(t, []string{"A", "B", "C"}, output.sortTags(tt.msg))
		})
	}
}

func TestKafkaOutput_encodeToInfluxLineProtocol(t *testing.T) {
	tests := []struct {
		name string
		msg  message.KafkaNormalizationMessage
		want string
	}{
		{
			name: "TestKafkaOutput_encodeToInfluxLineProtocol",
			msg: message.KafkaNormalizationMessage{
				Measurement: "normalization",
				Tags: map[string]string{
					"source":         "XR-1",
					"interface_name": "GigabitEthernet0/0/0/0",
				},
				Fields: map[string]float64{
					"normalized_unidir_link_delay": 0.00020253164556962027,
				},
			},
			want: "normalization,interface_name=GigabitEthernet0/0/0/0,source=XR-1 normalized_unidir_link_delay=0.00020253164556962027",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := config.KafkaOutputConfig{
				Broker: "localhost:9092",
				Topic:  "test",
			}
			output := NewKafkaOutput(config, make(chan message.Command), make(chan message.Result))
			encoder, shouldReturn := output.encodeToInfluxLineProtocol(tt.msg)
			assert.False(t, shouldReturn)
			assert.NotNil(t, encoder)
			encoderString := string(encoder.Bytes())
			assert.Contains(t, encoderString, tt.want)
		})
	}
}

func TestKafkaOutput_publishNotificationMessage(t *testing.T) {
	tests := []struct {
		name          string
		msg           message.KafkaNormalizationMessage
		wantEncodeErr bool
	}{
		{
			name: "TestKafkaOutput_publishNotificationMessage success",
			msg: message.KafkaNormalizationMessage{
				Measurement: "normalization",
				Fields: map[string]float64{
					"unidir_packet_loss_percentage_lower_fence": 0.0010077367045896624,
				},
			},
			wantEncodeErr: false,
		},
		{
			name: "TestKafkaOutput_publishNotificationMessage encoding err",
			msg: message.KafkaNormalizationMessage{
				Measurement: "",
				Fields: map[string]float64{
					"unidir_packet_loss_percentage_lower_fence": 0.0010077367045896624,
				},
			},
			wantEncodeErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := config.KafkaOutputConfig{
				Broker: "localhost:9092",
				Topic:  "test",
			}
			output := NewKafkaOutput(config, make(chan message.Command), make(chan message.Result))
			producer := NewKafkaProducerMock()
			if tt.wantEncodeErr {
				output.publishNotificationMessage(tt.msg)
				return
			}
			output.producer = producer
			wg := sync.WaitGroup{}
			wg.Add(1)
			go func() {
				<-producer.msgChan
				wg.Done()
			}()
			output.publishNotificationMessage(tt.msg)
			wg.Wait()
		})
	}
}

func TestKafkaOutput_publishEventNotifications(t *testing.T) {
	tests := []struct {
		name string
		msg  message.KafkaEventCommand
	}{
		{
			name: "TestKafkaOuput_publishEventNotifications",
			msg: message.KafkaEventCommand{
				Updates: []message.KafkaEventMessage{
					{
						TopicType: 1,
						Key:       "key",
						Id:        "id",
						Action:    "action",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := config.KafkaOutputConfig{
				Broker: "localhost:9092",
				Topic:  "test",
			}
			output := NewKafkaOutput(config, make(chan message.Command), make(chan message.Result))
			output.config.Type = "event-notification"
			producer := NewKafkaProducerMock()
			output.producer = producer
			wg := sync.WaitGroup{}
			wg.Add(1)
			go func() {
				<-producer.msgChan
				wg.Done()
			}()
			output.publishEventNotifications(tt.msg)
			wg.Wait()
		})
	}
}

func TestKafkaOutput_publishNormalizationMessages(t *testing.T) {
	tests := []struct {
		name string
		msg  message.KafkaNormalizationCommand
	}{
		{
			name: "TestKafkaOuput_pubhlishNormalizationMessages",
			msg: message.KafkaNormalizationCommand{
				Updates: []message.KafkaNormalizationMessage{
					{
						Measurement: "normalization",
						Tags: map[string]string{
							"source":         "XR-1",
							"interface_name": "GigabitEthernet0/0/0/0",
						},
						Fields: map[string]float64{
							"normalized_unidir_link_delay":      0.00020253164556962027,
							"normalized_unidir_delay_variation": 0.02028508771929825,
							"normalized_unidir_packet_loss":     0.00000046770242568585443,
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := config.KafkaOutputConfig{
				Broker: "localhost:9092",
				Topic:  "test",
			}
			output := NewKafkaOutput(config, make(chan message.Command), make(chan message.Result))
			producer := NewKafkaProducerMock()
			output.producer = producer
			output.config.Type = "normalization"
			wg := sync.WaitGroup{}
			wg.Add(1)
			go func() {
				<-producer.msgChan
				wg.Done()
			}()
			output.publishNormalizationMessages(tt.msg)
			wg.Wait()
		})
	}
}

func TestKafkaOutput_Start(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "TestKafkaOutput_Start",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := config.KafkaOutputConfig{
				Broker: "localhost:9092",
				Topic:  "test",
			}
			output := NewKafkaOutput(config, make(chan message.Command), make(chan message.Result))
			producer := NewKafkaProducerMock()
			output.producer = producer
			wg := sync.WaitGroup{}
			wg.Add(1)
			go func() {
				output.Start()
				wg.Done()
			}()
			output.commandChan <- message.KafkaEventCommand{}         // trigger event notification
			output.commandChan <- message.KafkaNormalizationCommand{} // trigger normalization
			output.commandChan <- message.BaseCommand{}               // unsupported command
			producer.errChan <- &sarama.ProducerError{}               // triger error log
			time.Sleep(100 * time.Millisecond)
			close(output.quitChan)
			wg.Wait()
		})
	}
}

func TestKafkaOutput_Stop(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "TestKafkaOutput_Stop success",
			wantErr: false,
		},
		{
			name:    "TestKafkaOutput_Stop success",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := config.KafkaOutputConfig{
				Broker: "localhost:9092",
				Topic:  "test",
			}
			output := NewKafkaOutput(config, make(chan message.Command), make(chan message.Result))
			producer := NewKafkaProducerMock()
			output.producer = producer
			wg := sync.WaitGroup{}
			wg.Add(1)
			go func() {
				output.Start()
				wg.Done()
			}()
			time.Sleep(100 * time.Millisecond)
			if tt.wantErr {
				producer.returnCloseErr = true
				assert.Error(t, output.Stop())
				return
			}
			assert.NoError(t, output.Stop())
			wg.Wait()
		})
	}
}
