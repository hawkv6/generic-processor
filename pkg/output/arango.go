package output

import (
	"context"
	"fmt"

	"github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/http"
	"github.com/hawkv6/generic-processor/pkg/config"
	"github.com/hawkv6/generic-processor/pkg/logging"
	"github.com/hawkv6/generic-processor/pkg/message"
	"github.com/jalapeno-api-gateway/jagw/pkg/arango"
	"github.com/sirupsen/logrus"
)

type ArangoOutput struct {
	log         *logrus.Entry
	config      config.ArangoOutputConfig
	client      driver.Client
	commandChan chan message.Command
	resultChan  chan message.Result
	quitChan    chan struct{}
}

func NewArangoOutput(config config.ArangoOutputConfig, commandChan chan message.Command, resultChan chan message.Result) *ArangoOutput {
	return &ArangoOutput{
		log:         logging.DefaultLogger.WithField("subsystem", Subsystem),
		config:      config,
		commandChan: commandChan,
		resultChan:  resultChan,
		quitChan:    make(chan struct{}),
	}
}

func (output *ArangoOutput) createNewConnection() (driver.Connection, error) {
	connection, err := http.NewConnection(http.ConnectionConfig{
		Endpoints: []string{output.config.URL},
	})
	if err != nil {
		return nil, fmt.Errorf("error creating ArangoDB connection: %v", err)
	}
	return connection, nil
}

func (output *ArangoOutput) createNewClient() error {
	connection, err := output.createNewConnection()
	if err != nil {
		return err
	}
	client, err := driver.NewClient(driver.ClientConfig{
		Connection:     connection,
		Authentication: driver.BasicAuthentication(output.config.Username, output.config.Password),
	})
	if err != nil {
		return fmt.Errorf("error creating ArangoDB client: %v", err)
	}
	output.client = client
	return nil
}

func (output *ArangoOutput) Init() error {
	output.log.Debugln("Initializing ArangoDB output: ", output.config.Name)
	if err := output.createNewClient(); err != nil {
		return err
	}
	return nil
}

func (output *ArangoOutput) getDatabase() (driver.Database, error) {
	ctx := context.Background()
	db, err := output.client.Database(ctx, output.config.DB)
	if err != nil {
		return nil, fmt.Errorf("error getting database '%s': %v", output.config.DB, err)
	}
	return db, nil
}

func (output *ArangoOutput) updateField(field interface{}, value float64) {
	switch fieldType := field.(type) {
	case *uint64:
		*fieldType = uint64(value)
	case *uint32:
		*fieldType = uint32(value)
	case *float32:
		*fieldType = float32(value)
	case *float64:
		*fieldType = value
	default:
		output.log.Errorf("Unsupported field type")
	}
}

type UpdateMessage struct {
	arango.LSLink
	NormalizedUnidirLinkDelay      float64 `json:"normalized_unidir_link_delay,omitempty"`
	NormalizedUnidirDelayVariation float64 `json:"normalized_unidir_delay_variation,omitempty"`
	NormalizedUnidirPacketLoss     float64 `json:"normalized_unidir_packet_loss,omitempty"`
	UnidirPacketLossPercentage     float64 `json:"undir_packet_loss_percentage,omitempty"`
}

func (output *ArangoOutput) processLsLinkDocument(ctx context.Context, cursor driver.Cursor, command message.ArangoUpdateCommand, updateMessages []UpdateMessage, keys []string, i int) (int, error) {
	updateMessage := UpdateMessage{}
	_, err := cursor.ReadDocument(ctx, &updateMessage)
	if err != nil {
		return i, err
	}
	arangoUpdate, ok := command.Updates[updateMessage.LocalLinkIP]
	if !ok {
		output.log.Errorf("No update found for local link IP: %s", updateMessage.LocalLinkIP)
		return i, nil
	}

	fields := map[string]interface{}{
		"unidir_link_delay":                 &updateMessage.UnidirLinkDelay,
		"unidir_link_delay_min_max[0]":      &updateMessage.UnidirLinkDelayMinMax[0],
		"unidir_link_delay_min_max[1]":      &updateMessage.UnidirLinkDelayMinMax[1],
		"unidir_delay_variation":            &updateMessage.UnidirDelayVariation,
		"unidir_packet_loss_percentage":     &updateMessage.UnidirPacketLossPercentage,
		"max_link_bw_kbps":                  &updateMessage.MaxLinkBWKbps,
		"unidir_available_bw":               &updateMessage.UnidirAvailableBW,
		"unidir_bw_utilization":             &updateMessage.UnidirBWUtilization,
		"normalized_unidir_link_delay":      &updateMessage.NormalizedUnidirLinkDelay,
		"normalized_unidir_delay_variation": &updateMessage.NormalizedUnidirDelayVariation,
		"normalized_unidir_packet_loss":     &updateMessage.NormalizedUnidirPacketLoss,
	}

	for j := 0; j < len(arangoUpdate.Fields); j++ {
		if field, ok := fields[arangoUpdate.Fields[j]]; ok {
			output.updateField(field, arangoUpdate.Values[j])
		}
	}
	updateMessage.UnidirAvailableBW = uint32(updateMessage.MaxLinkBWKbps) - updateMessage.UnidirBWUtilization

	updateMessages[i] = updateMessage
	keys[i] = updateMessage.Key
	return i + 1, nil
}

func (output *ArangoOutput) updateLsLink(ctx context.Context, cursor driver.Cursor, command message.ArangoUpdateCommand) ([]string, []UpdateMessage, error) {
	count := cursor.Count()
	keys := make([]string, count)
	lsLinks := make([]UpdateMessage, count)
	i := 0
	var err error
	for {
		i, err = output.processLsLinkDocument(ctx, cursor, command, lsLinks, keys, i)
		if driver.IsNoMoreDocuments(err) {
			return keys[:i], lsLinks[:i], nil
		} else if err != nil {
			output.log.Errorf("Unexpected error processing document: %v", err)
			continue
		}
	}
}

func (output *ArangoOutput) executeQuery(ctx context.Context, db driver.Database, command message.ArangoUpdateCommand) (driver.Cursor, error) {
	query := fmt.Sprintf("FOR d IN %s RETURN d", command.Collection)
	ctx = driver.WithQueryCount(ctx, true)
	cursor, err := db.Query(ctx, query, nil)
	if err != nil {
		output.log.Errorf("Error executing query: %v", err)
		return nil, err
	}
	return cursor, err
}

func (output *ArangoOutput) updateDocuments(ctx context.Context, db driver.Database, command message.ArangoUpdateCommand, keys []string, lsLinks []UpdateMessage) error {
	col, err := db.Collection(ctx, command.Collection)
	if err != nil {
		output.log.Errorf("Error getting collection: %v", err)
		return err
	}
	_, errSlice, err := col.UpdateDocuments(ctx, keys, lsLinks)
	if err != nil {
		output.log.Errorf("Error updating documents: %v %v", errSlice, err)
		return err
	}
	return nil
}

func (output *ArangoOutput) getTopicType(collection string) int {
	// from: https://github.com/cisco-open/jalapeno/blob/main/topology/dbclient/dbclient.go
	if collection == "ls_link" {
		return 9
	}
	output.log.Errorf("Unknown collection: %s", collection)
	return 0
}

func (output *ArangoOutput) processArangoUpdateCommand(command message.ArangoUpdateCommand) {
	output.log.Infof("Processing Arango update for collection: %s", command.Collection)
	db, err := output.getDatabase()
	if err != nil {
		output.log.Errorf("Error getting database: %v", err)
		return
	}
	ctx := context.Background()
	cursor, err := output.executeQuery(ctx, db, command)
	if err != nil {
		return
	}
	defer cursor.Close()

	if command.Collection == "ls_link" {
		keys, lsLinks, err := output.updateLsLink(ctx, cursor, command)
		if err != nil {
			output.log.Errorf("Error updating ls_link: %v", err)
			return
		}
		output.log.Infof("Updating %d documents in Arango", len(keys))
		if err := output.updateDocuments(ctx, db, command, keys, lsLinks); err != nil {
			output.log.Errorf("Error updating documents: %v", err)
			return
		}
		arangoEventNotificationMessage := message.ArangoEventNotificationMessage{
			EventMessages: make([]message.EventMessage, len(keys)),
		}

		arangoNormalizationMessage := message.ArangoNormalizationMessage{
			Measurement:           "normalization",
			NormalizationMessages: make([]message.NormalizationMessage, len(keys)+len(command.StatisticalData)),
		}

		offset := 0
		for key, value := range command.StatisticalData {
			arangoNormalizationMessage.NormalizationMessages[offset] = message.NormalizationMessage{
				Fields: map[string]float64{key: value},
			}
			offset++
		}

		for index, link := range lsLinks {
			if update, ok := command.Updates[link.LocalLinkIP]; ok {
				arangoNormalizationMessage.NormalizationMessages[offset+index] =
					message.NormalizationMessage{
						Tags: update.Tags,
						Fields: map[string]float64{
							"normalized_unidir_link_delay":      link.NormalizedUnidirLinkDelay,
							"normalized_unidir_delay_variation": link.NormalizedUnidirDelayVariation,
							"normalized_unidir_packet_loss":     link.NormalizedUnidirPacketLoss,
						},
					}
			} else {
				output.log.Errorf("No update found for local link IP: %s", link.LocalLinkIP)
				continue
			}

			arangoEventNotificationMessage.EventMessages[index] =
				message.EventMessage{
					Key:       link.Key,
					Id:        link.ID,
					TopicType: output.getTopicType(command.Collection),
				}
		}
		output.resultChan <- arangoEventNotificationMessage
		output.resultChan <- arangoNormalizationMessage
	} else {
		output.log.Errorf("Unknown collection: %s", command.Collection)
		return
	}
}

func (output *ArangoOutput) Start() {
	for {
		select {
		case command := <-output.commandChan:
			switch cmdType := command.(type) {
			case message.ArangoUpdateCommand:
				output.processArangoUpdateCommand(cmdType)
			default:
				output.log.Errorf("Unknown command type: %v", command)
			}
		case <-output.quitChan:
			return
		}
	}
}

func (output *ArangoOutput) Stop() error {
	output.log.Debugln("Stopping ArangoDB output: ", output.config.Name)
	close(output.quitChan)
	return nil
}
