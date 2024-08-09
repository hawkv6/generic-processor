package output

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/hawkv6/generic-processor/pkg/config"
	"github.com/hawkv6/generic-processor/pkg/message"
	"github.com/jalapeno-api-gateway/jagw/pkg/arango"
	"github.com/stretchr/testify/assert"
)

func TestNewArangoOutput(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "TestNewArangoOutput",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := config.ArangoOutputConfig{
				URL:      "http://arango:8528",
				DB:       "test",
				Username: "root",
				Password: "password",
			}
			assert.NotNil(t, NewArangoOutput(config, make(chan message.Command), make(chan message.Result)))
		})
	}
}

func TestArangoOutput_getHttpConnection(t *testing.T) {
	tests := []struct {
		name                string
		skipTlsVerification bool
	}{
		{
			name:                "TestArangoOutput_getHttpConnection",
			skipTlsVerification: false,
		},
		{
			name:                "TestArangoOutput_getHttpConnection skipTlsVerification",
			skipTlsVerification: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := config.ArangoOutputConfig{
				URL:                 "http://arango:8528",
				DB:                  "test",
				Username:            "root",
				Password:            "password",
				SkipTlsVerification: tt.skipTlsVerification,
			}
			output := NewArangoOutput(config, make(chan message.Command), make(chan message.Result))
			client, err := output.getHttpConnection()
			assert.NotNil(t, client)
			assert.NoError(t, err)
		})
	}
}

func TestArangoOutput_createNewClient(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "TestArangoOutput_createNewConnection",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := config.ArangoOutputConfig{
				URL: "",
				DB:  "test",
			}
			output := NewArangoOutput(config, make(chan message.Command), make(chan message.Result))
			err := output.createNewClient()
			assert.NoError(t, err)
		})
	}
}

func TestArangoOutput_Init(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "TestArangoOutput_Init",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := config.ArangoOutputConfig{
				URL:      "http://arango:8528",
				DB:       "test",
				Username: "root",
				Password: "password",
			}
			output := NewArangoOutput(config, make(chan message.Command), make(chan message.Result))
			err := output.Init()
			assert.NoError(t, err)
		})
	}
}

func TestArangoOutput_getDatabase(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "TestArangoOutput_getDatabase success",
			wantErr: false,
		},
		{
			name:    "TestArangoOutput_getDatabase error",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := config.ArangoOutputConfig{
				URL:      "http://arango:8528",
				DB:       "test",
				Username: "root",
				Password: "password",
			}
			output := NewArangoOutput(config, make(chan message.Command), make(chan message.Result))
			err := output.Init()
			assert.NoError(t, err)
			arangoMock := NewArangoClientMock()
			output.client = arangoMock
			if tt.wantErr {
				arangoMock.returnDbError = true
				_, err = output.getDatabase()
				assert.Error(t, err)
				return
			}
			_, err = output.getDatabase()
			assert.NoError(t, err)
		})
	}
}

func TestArangoOutput_updateField(t *testing.T) {
	fieldUint64 := uint64(10)
	fieldUint32 := uint32(10)
	fieldFloat32 := float32(10)
	fieldFloat64 := float64(10)
	field2Float64 := float64(0)

	tests := []struct {
		name          string
		fieldUint32   *uint32
		fieldUint64   *uint64
		fieldFloat32  *float32
		fieldFloat64  *float64
		field2Float64 *float64
		value         float64
	}{
		{
			name:          "TestArangoOutput_updateField ",
			fieldUint32:   &fieldUint32,
			fieldUint64:   &fieldUint64,
			fieldFloat32:  &fieldFloat32,
			fieldFloat64:  &fieldFloat64,
			field2Float64: &field2Float64,
			value:         10,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := NewArangoOutput(config.ArangoOutputConfig{}, make(chan message.Command), make(chan message.Result))
			output.updateField(tt.fieldUint32, tt.value)
			assert.Equal(t, *tt.fieldUint32, uint32(tt.value))
			output.updateField(tt.fieldUint64, tt.value)
			assert.Equal(t, *tt.fieldUint64, uint64(tt.value))
			output.updateField(tt.fieldFloat32, tt.value)
			assert.Equal(t, *tt.fieldFloat32, float32(tt.value))
			output.updateField(tt.fieldFloat64, tt.value)
			assert.Equal(t, *tt.fieldFloat64, tt.value)
			output.updateField(tt.fieldFloat64, tt.value)
			assert.Equal(t, *tt.fieldFloat64, tt.value)
			output.updateField(tt.field2Float64, 0.00001)
		})
	}
}

func TestArangoOutput_updateArangoFields(t *testing.T) {
	tests := []struct {
		name               string
		updateMessage      message.UpdateLinkMessage
		arangoUpdate       message.ArangoUpdate
		updateMessages     []message.UpdateLinkMessage
		initialLinkCounter int
	}{
		{
			name: "TestArangoOutput_updateArangoFields",
			updateMessage: message.UpdateLinkMessage{
				LSLink: arango.LSLink{
					UnidirLinkDelay:       10,
					UnidirLinkDelayMinMax: []uint32{10, 10},
					MaxLinkBWKbps:         10,
					UnidirDelayVariation:  10,
					UnidirResidualBW:      10,
					UnidirAvailableBW:     10,
					UnidirBWUtilization:   10,
				},
				NormalizedUnidirLinkDelay:      10,
				NormalizedUnidirDelayVariation: 10,
				NormalizedUnidirPacketLoss:     10,
				UnidirPacketLossPercentage:     10,
			},
			arangoUpdate: message.ArangoUpdate{
				Tags: map[string]string{"test": "test"},
				Fields: []string{
					"unidir_link_delay",
					"unidir_link_delay_min_max[0]",
					"unidir_link_delay_min_max[1]",
					"unidir_delay_variation",
					"unidir_packet_loss_percentage",
					"max_link_bw_kbps",
					"unidir_available_bw",
					"unidir_bw_utilization",
					"normalized_unidir_link_delay",
					"normalized_unidir_delay_variation",
					"normalized_unidir_packet_loss",
				},
				Values: []float64{20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20},
			},
			updateMessages:     make([]message.UpdateLinkMessage, 1),
			initialLinkCounter: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := NewArangoOutput(config.ArangoOutputConfig{}, make(chan message.Command), make(chan message.Result))
			keys := make([]string, 1)
			counter, err := output.updateArangoFields(tt.updateMessage, tt.arangoUpdate, tt.updateMessages, tt.initialLinkCounter, keys)
			assert.NoError(t, err)
			assert.Equal(t, counter, tt.initialLinkCounter+1)

			assert.Equal(t, tt.updateMessages[0].UnidirLinkDelay, uint32(tt.arangoUpdate.Values[0]))
			assert.Equal(t, tt.updateMessages[0].UnidirLinkDelayMinMax[0], uint32(tt.arangoUpdate.Values[1]))
			assert.Equal(t, tt.updateMessages[0].UnidirLinkDelayMinMax[1], uint32(tt.arangoUpdate.Values[2]))
			assert.Equal(t, tt.updateMessages[0].UnidirDelayVariation, uint32(tt.arangoUpdate.Values[3]))
			assert.Equal(t, tt.updateMessages[0].UnidirPacketLossPercentage, float64(tt.arangoUpdate.Values[4]))
			assert.Equal(t, tt.updateMessages[0].MaxLinkBWKbps, uint64(tt.arangoUpdate.Values[5]))
			assert.Equal(t, tt.updateMessages[0].UnidirAvailableBW, uint32(tt.arangoUpdate.Values[5])-uint32(tt.arangoUpdate.Values[7]))
			assert.Equal(t, tt.updateMessages[0].UnidirBWUtilization, uint32(tt.arangoUpdate.Values[7]))
			assert.Equal(t, tt.updateMessages[0].NormalizedUnidirLinkDelay, tt.arangoUpdate.Values[8])
			assert.Equal(t, tt.updateMessages[0].NormalizedUnidirDelayVariation, tt.arangoUpdate.Values[9])
			assert.Equal(t, tt.updateMessages[0].NormalizedUnidirPacketLoss, tt.arangoUpdate.Values[10])
		})
	}
}

func TestArangoOutput_processLsLinkDocument(t *testing.T) {
	localLinkIp := "2001:db8::1"
	updates := map[string]message.ArangoUpdate{
		localLinkIp: {
			Tags: map[string]string{"test": "test"},
			Fields: []string{
				"unidir_link_delay",
				"unidir_link_delay_min_max[0]",
				"unidir_link_delay_min_max[1]",
				"unidir_delay_variation",
				"unidir_packet_loss_percentage",
				"max_link_bw_kbps",
				"unidir_available_bw",
				"unidir_bw_utilization",
				"normalized_unidir_link_delay",
				"normalized_unidir_delay_variation",
				"normalized_unidir_packet_loss",
			},
			Values: []float64{20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20},
		},
	}
	updateMessage := message.UpdateLinkMessage{} // is filled in mock

	tests := []struct {
		name           string
		updates        map[string]message.ArangoUpdate
		updateMessage  message.UpdateLinkMessage
		updateMessages []message.UpdateLinkMessage
		keys           []string
		linkCounter    int
		wantErr        bool
		noUpdateFound  bool
	}{
		{
			name:           "TestArangoOutput_processLsLinkDocument ReadDocument error",
			updates:        updates,
			updateMessage:  updateMessage,
			updateMessages: make([]message.UpdateLinkMessage, 1),
			keys:           make([]string, 1),
			linkCounter:    0,
			wantErr:        true,
			noUpdateFound:  false,
		},
		{
			name:           "TestArangoOutput_processLsLinkDocument no update found for this local link ip",
			updates:        make(map[string]message.ArangoUpdate),
			updateMessage:  updateMessage,
			updateMessages: make([]message.UpdateLinkMessage, 1),
			keys:           make([]string, 1),
			linkCounter:    0,
			wantErr:        false,
			noUpdateFound:  true,
		},
		{
			name:           "TestArangoOutput_processLsLinkDocument success",
			updates:        updates,
			updateMessage:  updateMessage,
			updateMessages: make([]message.UpdateLinkMessage, 1),
			keys:           make([]string, 1),
			linkCounter:    0,
			wantErr:        false,
			noUpdateFound:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := NewArangoOutput(config.ArangoOutputConfig{}, make(chan message.Command), make(chan message.Result))
			ctx := context.Background()
			cursor := NewArangoCursorMock()
			if tt.wantErr {
				cursor.returnError = true
				counter, err := output.processLsLinkDocument(ctx, cursor, tt.updates, tt.updateMessages, tt.keys, tt.linkCounter)
				assert.Equal(t, counter, tt.linkCounter)
				assert.Error(t, err)
				return
			}
			if tt.noUpdateFound {
				counter, err := output.processLsLinkDocument(ctx, cursor, tt.updates, tt.updateMessages, tt.keys, tt.linkCounter)
				assert.Equal(t, counter, tt.linkCounter)
				assert.NoError(t, err)
				return
			}
			counter, err := output.processLsLinkDocument(ctx, cursor, tt.updates, tt.updateMessages, tt.keys, tt.linkCounter)
			assert.Equal(t, counter, tt.linkCounter+1)
			assert.NoError(t, err)
			assert.NotNil(t, tt.updateMessages[0])
			assert.Equal(t, tt.updateMessages[0].LocalLinkIP, localLinkIp)
			assert.Equal(t, tt.updateMessages[0].UnidirLinkDelay, uint32(tt.updates[localLinkIp].Values[0]))
			assert.Equal(t, tt.updateMessages[0].UnidirLinkDelayMinMax[0], uint32(tt.updates[localLinkIp].Values[1]))
			assert.Equal(t, tt.updateMessages[0].UnidirLinkDelayMinMax[1], uint32(tt.updates[localLinkIp].Values[2]))
			assert.Equal(t, tt.updateMessages[0].UnidirDelayVariation, uint32(tt.updates[localLinkIp].Values[3]))
			assert.Equal(t, tt.updateMessages[0].UnidirPacketLossPercentage, float64(tt.updates[localLinkIp].Values[4]))
			assert.Equal(t, tt.updateMessages[0].MaxLinkBWKbps, uint64(tt.updates[localLinkIp].Values[5]))
			assert.Equal(t, tt.updateMessages[0].UnidirAvailableBW, uint32(tt.updates[localLinkIp].Values[5])-uint32(tt.updates[localLinkIp].Values[7]))
			assert.Equal(t, tt.updateMessages[0].UnidirBWUtilization, uint32(tt.updates[localLinkIp].Values[7]))
			assert.Equal(t, tt.updateMessages[0].NormalizedUnidirLinkDelay, tt.updates[localLinkIp].Values[8])
			assert.Equal(t, tt.updateMessages[0].NormalizedUnidirDelayVariation, tt.updates[localLinkIp].Values[9])
			assert.Equal(t, tt.updateMessages[0].NormalizedUnidirPacketLoss, tt.updates[localLinkIp].Values[10])
		})
	}
}

func TestArangoOutput_updateLsLink(t *testing.T) {
	localLinkIp := "2001:db8::1"
	updates := map[string]message.ArangoUpdate{
		localLinkIp: {
			Tags: map[string]string{"test": "test"},
			Fields: []string{
				"unidir_link_delay",
				"unidir_link_delay_min_max[0]",
				"unidir_link_delay_min_max[1]",
				"unidir_delay_variation",
				"unidir_packet_loss_percentage",
				"max_link_bw_kbps",
				"unidir_available_bw",
				"unidir_bw_utilization",
				"normalized_unidir_link_delay",
				"normalized_unidir_delay_variation",
				"normalized_unidir_packet_loss",
			},
			Values: []float64{20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20},
		},
	}
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "TestArangoOutput_updateLsLink success",
			wantErr: false,
		},
		{
			name:    "TestArangoOutput_updateLsLink error",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := NewArangoOutput(config.ArangoOutputConfig{}, make(chan message.Command), make(chan message.Result))
			ctx := context.Background()
			cursor := NewArangoCursorMock()
			if tt.wantErr {
				cursor.returnError = true
				keys, lsLinks, err := output.updateLsLink(ctx, cursor, updates)
				assert.Error(t, err)
				assert.Nil(t, lsLinks)
				assert.Nil(t, keys)
				return
			}
			cursor.countReturnValue = 1
			keys, lsLinks, err := output.updateLsLink(ctx, cursor, updates)
			assert.NoError(t, err)
			assert.Equal(t, keys[0], localLinkIp) // for test purpose the cursor sets the key to local link ip
			assert.NotNil(t, lsLinks)
		})
	}
}

func TestArangoOutput_executeQuery(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "TestArangoOutput_executeQuery success",
			wantErr: false,
		},
		{
			name:    "TestArangoOutput_executeQuery error",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := NewArangoOutput(config.ArangoOutputConfig{}, make(chan message.Command), make(chan message.Result))
			ctx := context.Background()
			db := NewArangoDatabaseMock()
			if tt.wantErr {
				db.returnQueryError = true
				_, err := output.executeQuery(ctx, db, "test")
				assert.Error(t, err)
				return
			}
			_, err := output.executeQuery(ctx, db, "test")
			assert.NoError(t, err)
		})
	}
}

func TestArangoOutput_updateDocuments(t *testing.T) {
	tests := []struct {
		name          string
		wantDbErr     bool
		wantUpdateErr bool
	}{
		{
			name:          "TestArangoOutput_updateDocuments success",
			wantDbErr:     false,
			wantUpdateErr: false,
		},
		{
			name:          "TestArangoOutput_updateDocuments Database Collection error",
			wantDbErr:     true,
			wantUpdateErr: false,
		},
		{
			name:          "TestArangoOutput_updateDocuments Collection UpdateDocuments error",
			wantDbErr:     false,
			wantUpdateErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := NewArangoOutput(config.ArangoOutputConfig{}, make(chan message.Command), make(chan message.Result))
			ctx := context.Background()
			db := NewArangoDatabaseMock()
			if tt.wantDbErr {
				db.returnCollectionError = true
				err := output.updateDocuments(ctx, db, "test", []string{"test"}, []message.UpdateLinkMessage{})
				assert.Error(t, err)
				return
			}
			if tt.wantUpdateErr {
				db.returnUpdateDocumentsError = true
				err := output.updateDocuments(ctx, db, "test", []string{"test"}, []message.UpdateLinkMessage{})
				assert.Error(t, err)
				return
			}
			err := output.updateDocuments(ctx, db, "test", []string{"test"}, []message.UpdateLinkMessage{})
			assert.NoError(t, err)
		})
	}
}

func TestArangoOutput_getTopicType(t *testing.T) {
	tests := []struct {
		name       string
		collection string
		want       int
	}{
		{
			name:       "TestArangoOutput_getTopicType",
			collection: "ls_link",
			want:       9,
		},
		{
			name:       "TestArangoOutput_getTopicType unknown collection",
			collection: "unknown",
			want:       0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := NewArangoOutput(config.ArangoOutputConfig{}, make(chan message.Command), make(chan message.Result))
			assert.Equal(t, output.getTopicType(tt.collection), tt.want)
		})
	}
}

func TestArangoOutput_getDatabaseElements(t *testing.T) {
	tests := []struct {
		name         string
		wantDbErr    bool
		wantQueryErr bool
	}{
		{
			name:      "TestArangoOutput_getDatabaseElements DB error",
			wantDbErr: true,
		},
		{
			name:         "TeltArangoOutput_getDatabaseElements Query error",
			wantDbErr:    false,
			wantQueryErr: true,
		},
		{
			name:         "TeltArangoOutput_getDatabaseElements success",
			wantDbErr:    false,
			wantQueryErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := NewArangoOutput(config.ArangoOutputConfig{}, make(chan message.Command), make(chan message.Result))
			clientMock := NewArangoClientMock()
			output.client = clientMock
			if tt.wantDbErr {
				clientMock.returnDbError = true
				driver, context, cursor, shouldReturn := output.getDatabaseElements("test")
				assert.Nil(t, driver)
				assert.Nil(t, context)
				assert.Nil(t, cursor)
				assert.True(t, shouldReturn)
				return
			}
			if tt.wantQueryErr {
				clientMock.setQueryError = true
				driver, context, cursor, shouldReturn := output.getDatabaseElements("test")
				assert.Nil(t, driver)
				assert.Nil(t, context)
				assert.Nil(t, cursor)
				assert.True(t, shouldReturn)
				return
			}
			driver, context, cursor, shouldReturn := output.getDatabaseElements("test")
			assert.NotNil(t, driver)
			assert.NotNil(t, context)
			assert.NotNil(t, cursor)
			assert.False(t, shouldReturn)
		})
	}
}

func TestArangoOutput_updateLsLinkDocuments(t *testing.T) {
	localLinkIp := "2001:db8::1"
	command := message.ArangoUpdateCommand{
		Collection: "ls_link",
		Updates: map[string]message.ArangoUpdate{
			localLinkIp: {
				Tags: map[string]string{"test": "test"},
				Fields: []string{
					"unidir_link_delay",
					"unidir_link_delay_min_max[0]",
					"unidir_link_delay_min_max[1]",
					"unidir_delay_variation",
					"unidir_packet_loss_percentage",
					"max_link_bw_kbps",
					"unidir_available_bw",
					"unidir_bw_utilization",
					"normalized_unidir_link_delay",
					"normalized_unidir_delay_variation",
					"normalized_unidir_packet_loss",
				},
				Values: []float64{20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20},
			},
		},
	}
	tests := []struct {
		name                   string
		wantUpdateLsLinkErr    bool
		wantUpdateDocumentsErr bool
	}{
		{
			name:                   "TestArangoOutput_updatesLsLinkDocuments UpdateLsLink error",
			wantUpdateLsLinkErr:    true,
			wantUpdateDocumentsErr: false,
		},
		{
			name:                   "TestArangoOutput_updatesLsLinkDocuments error",
			wantUpdateLsLinkErr:    false,
			wantUpdateDocumentsErr: true,
		},
		{
			name:                   "TestArangoOutput_updatesLsLinkDocuments success",
			wantUpdateLsLinkErr:    false,
			wantUpdateDocumentsErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := NewArangoOutput(config.ArangoOutputConfig{}, make(chan message.Command), make(chan message.Result))
			ctx := context.Background()
			cursor := NewArangoCursorMock()
			db := NewArangoDatabaseMock()
			if tt.wantUpdateLsLinkErr {
				cursor.returnError = true
				keys, lsLinks, shouldReturn := output.updateLsLinkDocuments(ctx, cursor, command, db)
				assert.True(t, shouldReturn)
				assert.Nil(t, keys)
				assert.Nil(t, lsLinks)
				return
			}
			if tt.wantUpdateDocumentsErr {
				db.returnUpdateDocumentsError = true
				cursor.countReturnValue = 1
				keys, lsLinks, shouldReturn := output.updateLsLinkDocuments(ctx, cursor, command, db)
				assert.True(t, shouldReturn)
				assert.Nil(t, keys)
				assert.Nil(t, lsLinks)
				return
			}
			cursor.countReturnValue = 1
			keys, lsLinks, shouldReturn := output.updateLsLinkDocuments(ctx, cursor, command, db)
			assert.False(t, shouldReturn)
			assert.NotNil(t, keys)
			assert.NotNil(t, lsLinks)
		})
	}
}

func TestArangoOutput_getMessagesForFurtherProcessing(t *testing.T) {
	command := message.ArangoUpdateCommand{
		Collection: "ls_link",
		StatisticalData: map[string]float64{
			"unidir_link_delay_q3":                      2994,
			"unidir_packet_loss_percentage_q3":          2.00081046649455,
			"unidir_link_delay_upper_fence":             4485,
			"unidir_delay_variation_lower_fence":        6.482758620689655,
			"unidir_packet_loss_percentage_upper_fence": 3.006088581661236,
			"unidir_link_delay_q1":                      2000,
			"unidir_link_delay_iqr":                     994,
			"unidir_delay_variation_q1":                 8.89655172413793,
			"unidir_packet_loss_percentage_lower_fence": 0.0010081349220176184,
			"unidir_link_delay_lower_fence":             1997,
			"unidir_delay_variation_q3":                 64.72413793103448,
			"unidir_delay_variation_iqr":                55.82758620689655,
			"unidir_delay_variation_upper_fence":        148.4655172413793,
			"unidir_packet_loss_percentage_q1":          0.0010094377076123096,
			"unidir_packet_loss_percentage_iqr":         1.9998010287869377,
		},
	}

	tests := []struct {
		name    string
		wantErr bool
		keys    []string
		command message.ArangoUpdateCommand
	}{
		{
			name:    "TestArangoOutput_getMessagesForFurtherProcessing",
			keys:    []string{"2001:db8::1"},
			command: command,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := NewArangoOutput(config.ArangoOutputConfig{}, make(chan message.Command), make(chan message.Result))
			arangoEventNotificationMessage, arangoNormalizationMessage := output.getMessagesForFurtherProcessing(tt.keys, tt.command)
			assert.NotNil(t, arangoEventNotificationMessage)
			assert.NotNil(t, arangoNormalizationMessage)
			assert.Len(t, arangoNormalizationMessage.NormalizationMessages, len(tt.keys)+len(command.StatisticalData))
		})
	}
}

func TestArangoOutput_addStatisticalDataToNormalizationMessages(t *testing.T) {
	command := message.ArangoUpdateCommand{
		Collection: "ls_link",
		StatisticalData: map[string]float64{
			"unidir_link_delay_q3":                      2994,
			"unidir_packet_loss_percentage_q3":          2.00081046649455,
			"unidir_link_delay_upper_fence":             4485,
			"unidir_delay_variation_lower_fence":        6.482758620689655,
			"unidir_packet_loss_percentage_upper_fence": 3.006088581661236,
			"unidir_link_delay_q1":                      2000,
			"unidir_link_delay_iqr":                     994,
			"unidir_delay_variation_q1":                 8.89655172413793,
			"unidir_packet_loss_percentage_lower_fence": 0.0010081349220176184,
			"unidir_link_delay_lower_fence":             1997,
			"unidir_delay_variation_q3":                 64.72413793103448,
			"unidir_delay_variation_iqr":                55.82758620689655,
			"unidir_delay_variation_upper_fence":        148.4655172413793,
			"unidir_packet_loss_percentage_q1":          0.0010094377076123096,
			"unidir_packet_loss_percentage_iqr":         1.9998010287869377,
		},
	}
	tests := []struct {
		name string
	}{
		{
			name: "TestArangoOutput_addStatisticalDataToNormalizationMessages",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := NewArangoOutput(config.ArangoOutputConfig{}, make(chan message.Command), make(chan message.Result))
			arangoNormalizationMessage := message.ArangoNormalizationMessage{
				NormalizationMessages: make([]message.NormalizationMessage, len(command.StatisticalData)),
			}
			output.addStatisticalDataToNormalizationMessages(command, arangoNormalizationMessage)
			assert.Len(t, arangoNormalizationMessage.NormalizationMessages, len(command.StatisticalData))
		})
	}
}

func TestArangoOutput_addNormalizedValuesToNormalizationMessage(t *testing.T) {
	localLinkIp := "2001:db8::1"
	command := message.ArangoUpdateCommand{
		Collection: "ls_link",
		StatisticalData: map[string]float64{
			"unidir_link_delay_q3":                      2994,
			"unidir_packet_loss_percentage_q3":          2.00081046649455,
			"unidir_link_delay_upper_fence":             4485,
			"unidir_delay_variation_lower_fence":        6.482758620689655,
			"unidir_packet_loss_percentage_upper_fence": 3.006088581661236,
			"unidir_link_delay_q1":                      2000,
			"unidir_link_delay_iqr":                     994,
			"unidir_delay_variation_q1":                 8.89655172413793,
			"unidir_packet_loss_percentage_lower_fence": 0.0010081349220176184,
			"unidir_link_delay_lower_fence":             1997,
			"unidir_delay_variation_q3":                 64.72413793103448,
			"unidir_delay_variation_iqr":                55.82758620689655,
			"unidir_delay_variation_upper_fence":        148.4655172413793,
			"unidir_packet_loss_percentage_q1":          0.0010094377076123096,
			"unidir_packet_loss_percentage_iqr":         1.9998010287869377,
		},
		Updates: map[string]message.ArangoUpdate{
			localLinkIp: {
				Tags: map[string]string{"test": "test"},
				Fields: []string{
					"unidir_link_delay",
					"unidir_link_delay_min_max[0]",
					"unidir_link_delay_min_max[1]",
					"unidir_delay_variation",
					"unidir_packet_loss_percentage",
					"max_link_bw_kbps",
					"unidir_available_bw",
					"unidir_bw_utilization",
					"normalized_unidir_link_delay",
					"normalized_unidir_delay_variation",
					"normalized_unidir_packet_loss",
				},
				Values: []float64{20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20},
			},
		},
	}
	updatedLink := message.UpdateLinkMessage{
		LSLink: arango.LSLink{
			LocalLinkIP:           localLinkIp,
			UnidirLinkDelay:       20,
			UnidirLinkDelayMinMax: []uint32{20, 20},
			MaxLinkBWKbps:         20,
			UnidirDelayVariation:  20,
			UnidirResidualBW:      20,
			UnidirAvailableBW:     20,
			UnidirBWUtilization:   20,
		},
		NormalizedUnidirLinkDelay:      20,
		NormalizedUnidirDelayVariation: 20,
		NormalizedUnidirPacketLoss:     20,
		UnidirPacketLossPercentage:     20,
	}
	tests := []struct {
		name          string
		noUpdateFound bool
	}{
		{
			name:          "TestArangoOutput_addNormalizedLinkValuesToNormalizationMessages",
			noUpdateFound: false,
		},
		{
			name:          "TestArangoOutput_addNormalizedLinkValuesToNormalizationMessages no update found",
			noUpdateFound: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := NewArangoOutput(config.ArangoOutputConfig{}, make(chan message.Command), make(chan message.Result))
			offset := len(command.StatisticalData)
			arangoNormalizationMessage := message.ArangoNormalizationMessage{
				NormalizationMessages: make([]message.NormalizationMessage, offset+1),
			}
			if tt.noUpdateFound {
				updatedLink.LocalLinkIP = ""
				output.addNormalizedValuesToNormalizationMessage(command, updatedLink, arangoNormalizationMessage, offset, 0)
				assert.NotEqual(t, arangoNormalizationMessage.NormalizationMessages[offset].Fields["normalized_unidir_link_delay"], updatedLink.NormalizedUnidirLinkDelay)
				assert.NotEqual(t, arangoNormalizationMessage.NormalizationMessages[offset].Fields["normalized_unidir_delay_variation"], updatedLink.NormalizedUnidirDelayVariation)
				assert.NotEqual(t, arangoNormalizationMessage.NormalizationMessages[offset].Fields["normalized_unidir_packet_loss"], updatedLink.NormalizedUnidirPacketLoss)
				return
			}
			output.addNormalizedValuesToNormalizationMessage(command, updatedLink, arangoNormalizationMessage, offset, 0)
			assert.Equal(t, arangoNormalizationMessage.NormalizationMessages[offset].Fields["normalized_unidir_link_delay"], updatedLink.NormalizedUnidirLinkDelay)
			assert.Equal(t, arangoNormalizationMessage.NormalizationMessages[offset].Fields["normalized_unidir_delay_variation"], updatedLink.NormalizedUnidirDelayVariation)
			assert.Equal(t, arangoNormalizationMessage.NormalizationMessages[offset].Fields["normalized_unidir_packet_loss"], updatedLink.NormalizedUnidirPacketLoss)
		})
	}
}

func Test_ArangoOutput_addEventMessageToNotificationMessage(t *testing.T) {
	localLinkIp := "2001:db8::1"
	updatedLink := message.UpdateLinkMessage{
		LSLink: arango.LSLink{
			Key:                   "key",
			ID:                    "id",
			LocalLinkIP:           localLinkIp,
			UnidirLinkDelay:       20,
			UnidirLinkDelayMinMax: []uint32{20, 20},
			MaxLinkBWKbps:         20,
			UnidirDelayVariation:  20,
			UnidirResidualBW:      20,
			UnidirAvailableBW:     20,
			UnidirBWUtilization:   20,
		},
		NormalizedUnidirLinkDelay:      20,
		NormalizedUnidirDelayVariation: 20,
		NormalizedUnidirPacketLoss:     20,
		UnidirPacketLossPercentage:     20,
	}
	tests := []struct {
		name string
	}{
		{
			name: "Test_ArangoOutput_addEventMessageToNotificationMessage",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := NewArangoOutput(config.ArangoOutputConfig{}, make(chan message.Command), make(chan message.Result))
			notificationMessage := message.ArangoEventNotificationMessage{}
			notificationMessage.EventMessages = make([]message.EventMessage, 1)
			index := 0
			output.addEventMessageToNotificationMessage(notificationMessage, index, updatedLink, "test")
			assert.Len(t, notificationMessage.EventMessages, 1)
			assert.Equal(t, notificationMessage.EventMessages[index].Key, updatedLink.LSLink.Key)
			assert.Equal(t, notificationMessage.EventMessages[index].Id, updatedLink.LSLink.ID)
		})
	}
}

func TestArangoOutput_addLsLinkDataToMessages(t *testing.T) {
	localLinkIp := "2001:db8::1"
	updatedLink := message.UpdateLinkMessage{
		LSLink: arango.LSLink{
			Key:                   "key",
			ID:                    "id",
			LocalLinkIP:           localLinkIp,
			UnidirLinkDelay:       20,
			UnidirLinkDelayMinMax: []uint32{20, 20},
			MaxLinkBWKbps:         20,
			UnidirDelayVariation:  20,
			UnidirResidualBW:      20,
			UnidirAvailableBW:     20,
			UnidirBWUtilization:   20,
		},
		NormalizedUnidirLinkDelay:      20,
		NormalizedUnidirDelayVariation: 20,
		NormalizedUnidirPacketLoss:     20,
		UnidirPacketLossPercentage:     20,
	}
	command := message.ArangoUpdateCommand{
		Collection: "ls_link",
		StatisticalData: map[string]float64{
			"unidir_link_delay_q3":                      2994,
			"unidir_packet_loss_percentage_q3":          2.00081046649455,
			"unidir_link_delay_upper_fence":             4485,
			"unidir_delay_variation_lower_fence":        6.482758620689655,
			"unidir_packet_loss_percentage_upper_fence": 3.006088581661236,
			"unidir_link_delay_q1":                      2000,
			"unidir_link_delay_iqr":                     994,
			"unidir_delay_variation_q1":                 8.89655172413793,
			"unidir_packet_loss_percentage_lower_fence": 0.0010081349220176184,
			"unidir_link_delay_lower_fence":             1997,
			"unidir_delay_variation_q3":                 64.72413793103448,
			"unidir_delay_variation_iqr":                55.82758620689655,
			"unidir_delay_variation_upper_fence":        148.4655172413793,
			"unidir_packet_loss_percentage_q1":          0.0010094377076123096,
			"unidir_packet_loss_percentage_iqr":         1.9998010287869377,
		},
		Updates: map[string]message.ArangoUpdate{
			localLinkIp: {
				Tags: map[string]string{"test": "test"},
				Fields: []string{
					"unidir_link_delay",
					"unidir_link_delay_min_max[0]",
					"unidir_link_delay_min_max[1]",
					"unidir_delay_variation",
					"unidir_packet_loss_percentage",
					"max_link_bw_kbps",
					"unidir_available_bw",
					"unidir_bw_utilization",
					"normalized_unidir_link_delay",
					"normalized_unidir_delay_variation",
					"normalized_unidir_packet_loss",
				},
				Values: []float64{20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20},
			},
		},
	}
	tests := []struct {
		name string
	}{
		{
			name: "TestArangoOutput_addLsLinkDataToMessages",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := NewArangoOutput(config.ArangoOutputConfig{}, make(chan message.Command), make(chan message.Result))
			arangoEventNotificationMessage, arangoNormalizationMessage := output.getMessagesForFurtherProcessing([]string{localLinkIp}, command)
			lsLinks := []message.UpdateLinkMessage{updatedLink}
			offset := len(command.StatisticalData)
			output.addLsLinkDataToMessages(lsLinks, command, arangoNormalizationMessage, arangoEventNotificationMessage, offset)
		})
	}
}

func TestArangoOutput_processArangoUpdateCommand(t *testing.T) {
	localLinkIp := "2001:db8::1"
	command := message.ArangoUpdateCommand{
		StatisticalData: map[string]float64{
			"unidir_link_delay_q3":                      2994,
			"unidir_packet_loss_percentage_q3":          2.00081046649455,
			"unidir_link_delay_upper_fence":             4485,
			"unidir_delay_variation_lower_fence":        6.482758620689655,
			"unidir_packet_loss_percentage_upper_fence": 3.006088581661236,
			"unidir_link_delay_q1":                      2000,
			"unidir_link_delay_iqr":                     994,
			"unidir_delay_variation_q1":                 8.89655172413793,
			"unidir_packet_loss_percentage_lower_fence": 0.0010081349220176184,
			"unidir_link_delay_lower_fence":             1997,
			"unidir_delay_variation_q3":                 64.72413793103448,
			"unidir_delay_variation_iqr":                55.82758620689655,
			"unidir_delay_variation_upper_fence":        148.4655172413793,
			"unidir_packet_loss_percentage_q1":          0.0010094377076123096,
			"unidir_packet_loss_percentage_iqr":         1.9998010287869377,
		},
		Updates: map[string]message.ArangoUpdate{
			localLinkIp: {
				Tags: map[string]string{"test": "test"},
				Fields: []string{
					"unidir_link_delay",
					"unidir_link_delay_min_max[0]",
					"unidir_link_delay_min_max[1]",
					"unidir_delay_variation",
					"unidir_packet_loss_percentage",
					"max_link_bw_kbps",
					"unidir_available_bw",
					"unidir_bw_utilization",
					"normalized_unidir_link_delay",
					"normalized_unidir_delay_variation",
					"normalized_unidir_packet_loss",
				},
				Values: []float64{20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20},
			},
		},
	}
	tests := []struct {
		name           string
		collectionName string
		wantDBErr      bool
		wantUpdateErr  bool
	}{
		{
			name:           "TestArangoOutput_processArangoUpdateCommand not implemented collection",
			collectionName: "not implemented",
			wantDBErr:      false,
			wantUpdateErr:  false,
		},
		{
			name:           "TestArangoOutput_processArangoUpdateCommand db error",
			collectionName: "ls_link",
			wantDBErr:      true,
			wantUpdateErr:  false,
		},
		{
			name:           "TestArangoOutput_processArangoUpdateCommand db error",
			collectionName: "ls_link",
			wantDBErr:      false,
			wantUpdateErr:  true,
		},
		{
			name:           "TestArangoOutput_processArangoUpdateCommand success",
			collectionName: "ls_link",
			wantDBErr:      false,
			wantUpdateErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := NewArangoOutput(config.ArangoOutputConfig{}, make(chan message.Command), make(chan message.Result))
			dbClient := NewArangoClientMock()
			output.client = dbClient
			command.Collection = tt.collectionName
			if tt.collectionName != "ls_link" {
				output.processArangoUpdateCommand(command)
				return
			}
			if tt.wantDBErr {
				dbClient.returnDbError = true
				output.processArangoUpdateCommand(command)
				return
			}
			dbMock, err := dbClient.Database(context.Background(), "test")
			assert.NoError(t, err)
			cursor, err := dbMock.Query(context.Background(), "test", nil)
			assert.NoError(t, err)
			cursor.(*ArangoCursorMock).internalCount = 1
			if tt.wantUpdateErr {
				dbMock.(*ArangoDatabaseMock).returnCollectionError = true
				output.processArangoUpdateCommand(command)
				return
			}
			wg := sync.WaitGroup{}
			wg.Add(1)
			go func() {
				<-output.resultChan
				<-output.resultChan
				wg.Done()
			}()
			output.processArangoUpdateCommand(command)
			wg.Wait()
		})
	}
}

func TestArangoOutput_Start(t *testing.T) {
	localLinkIp := "2001:db8::1"
	command := message.ArangoUpdateCommand{
		StatisticalData: map[string]float64{
			"unidir_link_delay_q3":                      2994,
			"unidir_packet_loss_percentage_q3":          2.00081046649455,
			"unidir_link_delay_upper_fence":             4485,
			"unidir_delay_variation_lower_fence":        6.482758620689655,
			"unidir_packet_loss_percentage_upper_fence": 3.006088581661236,
			"unidir_link_delay_q1":                      2000,
			"unidir_link_delay_iqr":                     994,
			"unidir_delay_variation_q1":                 8.89655172413793,
			"unidir_packet_loss_percentage_lower_fence": 0.0010081349220176184,
			"unidir_link_delay_lower_fence":             1997,
			"unidir_delay_variation_q3":                 64.72413793103448,
			"unidir_delay_variation_iqr":                55.82758620689655,
			"unidir_delay_variation_upper_fence":        148.4655172413793,
			"unidir_packet_loss_percentage_q1":          0.0010094377076123096,
			"unidir_packet_loss_percentage_iqr":         1.9998010287869377,
		},
		Updates: map[string]message.ArangoUpdate{
			localLinkIp: {
				Tags: map[string]string{"test": "test"},
				Fields: []string{
					"unidir_link_delay",
					"unidir_link_delay_min_max[0]",
					"unidir_link_delay_min_max[1]",
					"unidir_delay_variation",
					"unidir_packet_loss_percentage",
					"max_link_bw_kbps",
					"unidir_available_bw",
					"unidir_bw_utilization",
					"normalized_unidir_link_delay",
					"normalized_unidir_delay_variation",
					"normalized_unidir_packet_loss",
				},
				Values: []float64{20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20},
			},
		},
	}
	tests := []struct {
		name string
	}{
		{
			name: "TestArangoOutput_Start valid command",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := NewArangoOutput(config.ArangoOutputConfig{}, make(chan message.Command), make(chan message.Result))
			wg := sync.WaitGroup{}
			wg.Add(1)
			go func() {
				output.Start()
				wg.Done()
			}()
			output.commandChan <- message.BaseCommand{} // unknown command
			output.commandChan <- command
			time.Sleep(100 * time.Millisecond)
			close(output.quitChan)
			wg.Wait()
		})
	}
}

func TestArangoOutput_Stop(t *testing.T) {
	localLinkIp := "2001:db8::1"
	command := message.ArangoUpdateCommand{
		StatisticalData: map[string]float64{
			"unidir_link_delay_q3":                      2994,
			"unidir_packet_loss_percentage_q3":          2.00081046649455,
			"unidir_link_delay_upper_fence":             4485,
			"unidir_delay_variation_lower_fence":        6.482758620689655,
			"unidir_packet_loss_percentage_upper_fence": 3.006088581661236,
			"unidir_link_delay_q1":                      2000,
			"unidir_link_delay_iqr":                     994,
			"unidir_delay_variation_q1":                 8.89655172413793,
			"unidir_packet_loss_percentage_lower_fence": 0.0010081349220176184,
			"unidir_link_delay_lower_fence":             1997,
			"unidir_delay_variation_q3":                 64.72413793103448,
			"unidir_delay_variation_iqr":                55.82758620689655,
			"unidir_delay_variation_upper_fence":        148.4655172413793,
			"unidir_packet_loss_percentage_q1":          0.0010094377076123096,
			"unidir_packet_loss_percentage_iqr":         1.9998010287869377,
		},
		Updates: map[string]message.ArangoUpdate{
			localLinkIp: {
				Tags: map[string]string{"test": "test"},
				Fields: []string{
					"unidir_link_delay",
					"unidir_link_delay_min_max[0]",
					"unidir_link_delay_min_max[1]",
					"unidir_delay_variation",
					"unidir_packet_loss_percentage",
					"max_link_bw_kbps",
					"unidir_available_bw",
					"unidir_bw_utilization",
					"normalized_unidir_link_delay",
					"normalized_unidir_delay_variation",
					"normalized_unidir_packet_loss",
				},
				Values: []float64{20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20},
			},
		},
	}
	tests := []struct {
		name string
	}{
		{
			name: "TestArangoOutput_Start valid command",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := NewArangoOutput(config.ArangoOutputConfig{}, make(chan message.Command), make(chan message.Result))
			wg := sync.WaitGroup{}
			wg.Add(1)
			go func() {
				output.Start()
				wg.Done()
			}()
			output.commandChan <- message.BaseCommand{} // unknown command
			output.commandChan <- command
			time.Sleep(100 * time.Millisecond)
			assert.NoError(t, output.Stop())
			wg.Wait()
		})
	}
}
