package output

import (
	"context"
	"crypto/tls"
	"fmt"
	"strconv"

	"github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/http"
	"github.com/hawkv6/generic-processor/pkg/config"
	"github.com/hawkv6/generic-processor/pkg/logging"
	"github.com/hawkv6/generic-processor/pkg/message"
	"github.com/sirupsen/logrus"
)

type ArangoOutput struct {
	log         *logrus.Entry
	config      config.ArangoOutputConfig
	client      ArangoClient
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

func (output *ArangoOutput) getHttpConnection() (driver.Connection, error) {
	if output.config.SkipTlsVerification {
		return http.NewConnection(http.ConnectionConfig{
			Endpoints: []string{output.config.URL},
			TLSConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		})
	}
	return http.NewConnection(http.ConnectionConfig{
		Endpoints: []string{output.config.URL},
	})
}

func (output *ArangoOutput) createNewClient() error {
	connection, err := output.getHttpConnection()
	if err != nil {
		return fmt.Errorf("error creating ArangoDB connection: %v", err)
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
		formattedValue := strconv.FormatFloat(value, 'f', 5, 64)
		parsedValue, _ := strconv.ParseFloat(formattedValue, 64)
		if parsedValue <= 0 {
			*fieldType = 0.00001
			return
		}
		*fieldType = parsedValue
	default:
		output.log.Errorf("Unsupported field type %t, ", fieldType)
	}
}

func (output *ArangoOutput) updateArangoFields(updateMessage message.UpdateLinkMessage, arangoUpdate message.ArangoUpdate, updateMessages []message.UpdateLinkMessage, linkCounter int, keys []string) (int, error) {
	fields := updateMessage.GetFields()
	for j := 0; j < len(arangoUpdate.Fields); j++ {
		if field, ok := fields[arangoUpdate.Fields[j]]; ok {
			output.updateField(field, arangoUpdate.Values[j])
		}
	}
	updateMessage.UnidirAvailableBW = uint32(updateMessage.MaxLinkBWKbps) - updateMessage.UnidirBWUtilization

	updateMessages[linkCounter] = updateMessage
	keys[linkCounter] = updateMessage.Key
	return linkCounter + 1, nil
}

func (output *ArangoOutput) processLsLinkDocument(ctx context.Context, cursor ArangoCursor, updates map[string]message.ArangoUpdate, updateMessages []message.UpdateLinkMessage, keys []string, linkCounter int) (int, error) {
	updateMessage := message.UpdateLinkMessage{}
	_, err := cursor.ReadDocument(ctx, &updateMessage)
	if err != nil {
		return linkCounter, err
	}
	arangoUpdate, ok := updates[updateMessage.LocalLinkIP]
	if !ok {
		output.log.Errorf("No update found for local link IP: %s", updateMessage.LocalLinkIP)
		return linkCounter, nil
	}
	return output.updateArangoFields(updateMessage, arangoUpdate, updateMessages, linkCounter, keys)
}

func (output *ArangoOutput) updateLsLink(ctx context.Context, cursor ArangoCursor, updates map[string]message.ArangoUpdate) ([]string, []message.UpdateLinkMessage, error) {
	count := cursor.Count()
	keys := make([]string, count)
	lsLinks := make([]message.UpdateLinkMessage, count)
	linkCounter := 0
	var err error
	for {
		linkCounter, err = output.processLsLinkDocument(ctx, cursor, updates, lsLinks, keys, linkCounter)
		if driver.IsNoMoreDocuments(err) {
			return keys[:linkCounter], lsLinks[:linkCounter], nil
		} else if err != nil {
			return nil, nil, fmt.Errorf("Unexpected error processing document: %v", err)
		}
	}
}

func (output *ArangoOutput) executeQuery(ctx context.Context, db ArangoDatabase, collection string) (driver.Cursor, error) {
	query := fmt.Sprintf("FOR d IN %s RETURN d", collection)
	ctx = driver.WithQueryCount(ctx, true)
	cursor, err := db.Query(ctx, query, nil)
	if err != nil {
		output.log.Errorf("Error executing query: %v", err)
		return nil, err
	}
	return cursor, err
}

func (output *ArangoOutput) updateDocuments(ctx context.Context, db ArangoDatabase, collection string, keys []string, lsLinks []message.UpdateLinkMessage) error {
	col, err := db.Collection(ctx, collection)
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

func (output *ArangoOutput) getDatabaseElements(collection string) (driver.Database, context.Context, driver.Cursor, bool) {
	db, err := output.getDatabase()
	if err != nil {
		output.log.Errorf("Error getting database: %v", err)
		return nil, nil, nil, true
	}
	ctx := context.Background()
	cursor, err := output.executeQuery(ctx, db, collection)
	if err != nil {
		output.log.Errorf("Error executing query: %v", err)
		return nil, nil, nil, true
	}
	return db, ctx, cursor, false
}

func (output *ArangoOutput) updateLsLinkDocuments(ctx context.Context, cursor driver.Cursor, command message.ArangoUpdateCommand, db driver.Database) ([]string, []message.UpdateLinkMessage, bool) {
	keys, lsLinks, err := output.updateLsLink(ctx, cursor, command.Updates)
	if err != nil {
		output.log.Errorf("Error updating ls_link: %v", err)
		return nil, nil, true
	}
	output.log.Infof("Updating %d documents in Arango DB", len(keys))
	if err := output.updateDocuments(ctx, db, command.Collection, keys, lsLinks); err != nil {
		output.log.Errorf("Error updating documents: %v", err)
		return nil, nil, true
	}
	return keys, lsLinks, false
}

func (*ArangoOutput) getMessagesForFurtherProcessing(keys []string, command message.ArangoUpdateCommand) (message.ArangoEventNotificationMessage, message.ArangoNormalizationMessage) {
	arangoEventNotificationMessage := message.ArangoEventNotificationMessage{
		EventMessages: make([]message.EventMessage, len(keys)),
	}
	arangoNormalizationMessage := message.ArangoNormalizationMessage{
		Measurement:           "normalization",
		NormalizationMessages: make([]message.NormalizationMessage, len(command.StatisticalData)+len(keys)),
	}
	return arangoEventNotificationMessage, arangoNormalizationMessage
}

func (*ArangoOutput) addStatisticalDataToNormalizationMessages(command message.ArangoUpdateCommand, arangoNormalizationMessage message.ArangoNormalizationMessage) int {
	offset := 0
	for key, value := range command.StatisticalData {
		arangoNormalizationMessage.NormalizationMessages[offset] = message.NormalizationMessage{
			Fields: map[string]float64{key: value},
		}
		offset++
	}
	return offset
}

func (output *ArangoOutput) addNormalizedValuesToNormalizationMessage(command message.ArangoUpdateCommand, updatedLink message.UpdateLinkMessage, arangoNormalizationMessage message.ArangoNormalizationMessage, offset int, index int) {
	if update, ok := command.Updates[updatedLink.LocalLinkIP]; ok {
		arangoNormalizationMessage.NormalizationMessages[offset+index] =
			message.NormalizationMessage{
				Tags: update.Tags,
				Fields: map[string]float64{
					"normalized_unidir_link_delay":      updatedLink.NormalizedUnidirLinkDelay,
					"normalized_unidir_delay_variation": updatedLink.NormalizedUnidirDelayVariation,
					"normalized_unidir_packet_loss":     updatedLink.NormalizedUnidirPacketLoss,
				},
			}
	} else {
		output.log.Errorf("No update found for local link IP: %s", updatedLink.LocalLinkIP)
		return
	}
}

func (output *ArangoOutput) addEventMessageToNotificationMessage(arangoEventNotificationMessage message.ArangoEventNotificationMessage, index int, updatedLink message.UpdateLinkMessage, collection string) {
	arangoEventNotificationMessage.EventMessages[index] =
		message.EventMessage{
			Key:       updatedLink.Key,
			Id:        updatedLink.ID,
			TopicType: output.getTopicType(collection),
		}
}

func (output *ArangoOutput) addLsLinkDataToMessages(lsLinks []message.UpdateLinkMessage, command message.ArangoUpdateCommand, arangoNormalizationMessage message.ArangoNormalizationMessage, arangoEventNotificationMessage message.ArangoEventNotificationMessage, offset int) {
	for index, link := range lsLinks {
		output.addNormalizedValuesToNormalizationMessage(command, link, arangoNormalizationMessage, offset, index)
		output.addEventMessageToNotificationMessage(arangoEventNotificationMessage, index, link, command.Collection)
	}
}

func (output *ArangoOutput) processArangoUpdateCommand(command message.ArangoUpdateCommand) {
	output.log.Infof("Processing Arango update for collection: %s", command.Collection)
	if command.Collection != "ls_link" {
		output.log.Errorf("Collection '%s' not yet implemented", command.Collection)
		return
	}

	db, ctx, cursor, shouldReturn := output.getDatabaseElements(command.Collection)
	if shouldReturn {
		return
	}
	defer cursor.Close()

	keys, lsLinks, shouldReturn := output.updateLsLinkDocuments(ctx, cursor, command, db)
	if shouldReturn {
		return
	}

	arangoEventNotificationMessage, arangoNormalizationMessage := output.getMessagesForFurtherProcessing(keys, command)

	offset := output.addStatisticalDataToNormalizationMessages(command, arangoNormalizationMessage)

	output.addLsLinkDataToMessages(lsLinks, command, arangoNormalizationMessage, arangoEventNotificationMessage, offset)
	output.resultChan <- arangoEventNotificationMessage
	output.resultChan <- arangoNormalizationMessage
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
