package processor

import (
	"sync"
	"testing"
	"time"

	"github.com/hawkv6/generic-processor/pkg/message"
	"github.com/stretchr/testify/assert"
)

func TestNewKafkaOpenConfigProcessor(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "Test NewKafkaOpenConfigProcessor",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotNil(t, NewKafkaOpenConfigProcessor())
		})
	}
}

func TestKafkaOpenConfigProcessor_processIpv6Message(t *testing.T) {
	testMessage := &message.IPv6Message{
		TelemetryMessage: message.TelemetryMessage{
			Tags: message.MessageTags{
				InterfaceName: "GigabitEthernet0/0/0/0",
				Source:        "router1",
			},
		},
		Fields: message.IPv6Fields{
			IPv6:         "2001:db8::1",
			Delete:       false,
			PrefixLength: 64,
		},
	}

	tests := []struct {
		name        string
		message     *message.IPv6Message
		addIpv6Addr bool
	}{
		{
			name:        "Test KafkaOpenConfigProcessor_processIpv6Message add message",
			message:     testMessage,
			addIpv6Addr: true,
		},
		{
			name:        "Test KafkaOpenConfigProcessor_processIpv6Message delete message",
			message:     testMessage,
			addIpv6Addr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := NewKafkaOpenConfigProcessor()
			assert.NotNil(t, processor)
			if !tt.addIpv6Addr {
				tt.message.Fields.Delete = true
				processor.activeIpv6Addresses[tt.message.Tags.Source+tt.message.Tags.InterfaceName] = tt.message.Fields.IPv6
			}
			processor.processIpv6Message(tt.message)
			if tt.addIpv6Addr {
				assert.Equal(t, tt.message.Fields.IPv6, processor.activeIpv6Addresses[tt.message.Tags.Source+tt.message.Tags.InterfaceName])
			} else {
				assert.Empty(t, processor.activeIpv6Addresses)
			}
		})
	}
}

func TestKafkaOpenConfigProcessor_handleInterfaceUp(t *testing.T) {
	testIpv6Address := "2001:db8::1"
	testMessage := &message.InterfaceStatusMessage{
		TelemetryMessage: message.TelemetryMessage{
			Tags: message.MessageTags{
				InterfaceName: "GigabitEthernet0/0/0/0",
				Source:        "router1",
			},
		},
		Fields: message.InterfaceStatusFields{
			AdminStatus: "UP",
		},
	}

	tests := []struct {
		name    string
		message *message.InterfaceStatusMessage
	}{
		{
			name:    "Test KafkaOpenConfigProcessor_handleInterfaceUp ",
			message: testMessage,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := NewKafkaOpenConfigProcessor()
			assert.NotNil(t, processor)
			processor.deactivatedIpv6Addresses[tt.message.Tags.Source+tt.message.Tags.InterfaceName] = testIpv6Address
			processor.handleInterfaceUp(tt.message)
			assert.Equal(t, testIpv6Address, processor.activeIpv6Addresses[tt.message.Tags.Source+tt.message.Tags.InterfaceName])
			assert.Empty(t, processor.deactivatedIpv6Addresses)
		})
	}
}

func TestKafkaOpenConfigProcessor_handleInterfaceDown(t *testing.T) {
	testIpv6Address := "2001:db8::1"
	testMessage := &message.InterfaceStatusMessage{
		TelemetryMessage: message.TelemetryMessage{
			Tags: message.MessageTags{
				InterfaceName: "GigabitEthernet0/0/0/0",
				Source:        "router1",
			},
		},
		Fields: message.InterfaceStatusFields{
			AdminStatus: "DOWN",
		},
	}

	tests := []struct {
		name    string
		message *message.InterfaceStatusMessage
	}{
		{
			name:    "Test KafkaOpenConfigProcessor_handleInterfaceDown ",
			message: testMessage,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := NewKafkaOpenConfigProcessor()
			assert.NotNil(t, processor)
			processor.activeIpv6Addresses[tt.message.Tags.Source+tt.message.Tags.InterfaceName] = testIpv6Address
			processor.handleInterfaceDown(tt.message)
			assert.Equal(t, testIpv6Address, processor.deactivatedIpv6Addresses[tt.message.Tags.Source+tt.message.Tags.InterfaceName])
			assert.Empty(t, processor.activeIpv6Addresses)
		})
	}
}

func TestKafkaOpenConfigProcessor_processInterfaceStatusMessage(t *testing.T) {
	testIpv6Address := "2001:db8::1"
	testMessage := &message.InterfaceStatusMessage{
		TelemetryMessage: message.TelemetryMessage{
			Tags: message.MessageTags{
				InterfaceName: "GigabitEthernet0/0/0/0",
				Source:        "router1",
			},
		},
		Fields: message.InterfaceStatusFields{
			AdminStatus: "UP",
		},
	}

	tests := []struct {
		name    string
		message *message.InterfaceStatusMessage
	}{
		{
			name:    "Test KafkaOpenConfigProcessor_processInterfaceStatusMessage UP",
			message: testMessage,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := NewKafkaOpenConfigProcessor()
			assert.NotNil(t, processor)
			processor.deactivatedIpv6Addresses[tt.message.Tags.Source+tt.message.Tags.InterfaceName] = testIpv6Address
			processor.processInterfaceStatusMessage(tt.message)
			assert.Equal(t, testIpv6Address, processor.activeIpv6Addresses[tt.message.Tags.Source+tt.message.Tags.InterfaceName])
			assert.Empty(t, processor.deactivatedIpv6Addresses)
			tt.message.Fields.AdminStatus = "DOWN"
			processor.processInterfaceStatusMessage(tt.message)
			assert.Empty(t, processor.activeIpv6Addresses)
			assert.Equal(t, testIpv6Address, processor.deactivatedIpv6Addresses[tt.message.Tags.Source+tt.message.Tags.InterfaceName])
		})
	}
}

func TestKafkaOpenConfigProcessor_Start(t *testing.T) {
	statusMessage := &message.InterfaceStatusMessage{
		TelemetryMessage: message.TelemetryMessage{
			Tags: message.MessageTags{
				InterfaceName: "GigabitEthernet0/0/0/0",
				Source:        "router1",
			},
		},
		Fields: message.InterfaceStatusFields{
			AdminStatus: "UP",
		},
	}
	ipv6Message := &message.IPv6Message{
		TelemetryMessage: message.TelemetryMessage{
			Tags: message.MessageTags{
				InterfaceName: "GigabitEthernet0/0/0/0",
				Source:        "router1",
			},
		},
		Fields: message.IPv6Fields{
			IPv6:         "2001:db8::1",
			Delete:       false,
			PrefixLength: 64,
		},
	}
	tests := []struct {
		name    string
		message message.TelemetryMessage
	}{
		{
			name: "Test KafkaOpenConfigProcessor_Start",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := NewKafkaOpenConfigProcessor()
			assert.NotNil(t, processor)
			commandChan := make(chan message.Command)
			resultChan := make(chan message.Result)
			wg := sync.WaitGroup{}
			wg.Add(2)
			go func() {
				<-commandChan
				wg.Done()
			}()
			go func() {
				processor.Start("test", commandChan, resultChan)
				wg.Done()
			}()
			resultChan <- ipv6Message
			time.Sleep(100 * time.Millisecond)
			resultChan <- statusMessage
			time.Sleep(100 * time.Millisecond)
			resultChan <- &message.BaseResultmessage{} // trigger default case
			time.Sleep(100 * time.Millisecond)
			close(processor.quitChan)
			wg.Wait()
		})
	}
}

func TestKafkaOpenConfigProcessor_GetLocalLinkIp(t *testing.T) {
	tests := []struct {
		name    string
		tags    map[string]string
		wantErr bool
	}{
		{
			name: "Test GetLocalLinkIp success",
			tags: map[string]string{
				"source":         "router1",
				"interface_name": "GigabitEthernet0/0/0/0",
			},
			wantErr: false,
		},
		{
			name: "Test GetLocalLinkIp other tags",
			tags: map[string]string{
				"source":         "router1",
				"other_tag":      "other_value",
				"interface_name": "GigabitEthernet0/0/0/0",
			},
			wantErr: false,
		},
		{
			name: "Test GetLocalLinkIp no source tag error",
			tags: map[string]string{
				"other_tag": "other_value",
			},
			wantErr: true,
		},
		{
			name: "Test GetLocalLinkIp no ipv6 found error",
			tags: map[string]string{
				"source":    "router1",
				"other_tag": "other_value",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := NewKafkaOpenConfigProcessor()
			assert.NotNil(t, processor)
			processor.activeIpv6Addresses[tt.tags["source"]+tt.tags["interface_name"]] = "2001:db8::1"
			if tt.wantErr {
				ipv6Address, err := processor.GetLocalLinkIp(tt.tags)
				assert.Empty(t, ipv6Address)
				assert.Error(t, err)
				return
			}
			ipv6Address, err := processor.GetLocalLinkIp(tt.tags)
			assert.NotEmpty(t, ipv6Address)
			assert.NoError(t, err)
		})
	}
}

func TestKafkaOpenConfigProcessor_Stop(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "Test KafkaOpenConfigProcessor_Stop",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := NewKafkaOpenConfigProcessor()
			assert.NotNil(t, processor)
			commandChan := make(chan message.Command)
			wg := sync.WaitGroup{}
			wg.Add(2)
			go func() {
				<-commandChan
				wg.Done()
			}()
			go func() {
				processor.Start("test", commandChan, make(chan message.Result))
				wg.Done()
			}()
			processor.Stop()
			wg.Wait()
		})
	}
}
