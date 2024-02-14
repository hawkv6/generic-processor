package output

import (
	"context"
	"encoding/json"
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

func (output *ArangoOutput) createNewConnection() (error, driver.Connection) {
	connection, err := http.NewConnection(http.ConnectionConfig{
		Endpoints: []string{output.config.URL},
	})
	if err != nil {
		return fmt.Errorf("error creating ArangoDB connection: %v", err), nil
	}
	return nil, connection
}

func (output *ArangoOutput) createNewClient() error {
	err, connection := output.createNewConnection()
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

func (output *ArangoOutput) getDatabase() (error, driver.Database) {
	ctx := context.Background()
	db, err := output.client.Database(ctx, output.config.DB)
	if err != nil {
		return fmt.Errorf("error getting database '%s': %v", output.config.DB, err), nil
	}
	return nil, db
}

func (output *ArangoOutput) updateField(field *uint32, value json.Number) {
	if intValue, err := value.Int64(); err == nil {
		*field = uint32(intValue)
	} else if floatValue, err := value.Float64(); err == nil {
		*field = uint32(floatValue)
	} else {
		output.log.Errorf("Failed to convert json.Number to int64 or float64")
	}
}

func (output *ArangoOutput) processLsLinkDocument(ctx context.Context, cursor driver.Cursor, command message.ArangoUpdateCommand, lsLinks []arango.LSLink, keys []string, i int) (int, error) {
	lsLink := arango.LSLink{}
	_, err := cursor.ReadDocument(ctx, &lsLink)
	if err != nil {
		return i, err
	}
	arangoUpdate, ok := command.Updates[lsLink.LocalLinkIP]
	if !ok {
		output.log.Errorf("No update found for local link IP: %s", lsLink.LocalLinkIP)
		return i, nil
	}
	if jsonNumber, ok := arangoUpdate.Value.(json.Number); ok {
		if arangoUpdate.Field == "unidir_link_delay" {
			output.updateField(&lsLink.UnidirLinkDelay, jsonNumber)
		} else if arangoUpdate.Field == "unidir_link_delay_min_max" {
			if arangoUpdate.Index == nil || *arangoUpdate.Index > 1 {
				output.log.Errorf("Unknown index: %v", arangoUpdate.Index)
				return i, nil
			}
			output.updateField(&lsLink.UnidirLinkDelayMinMax[*arangoUpdate.Index], jsonNumber)
		}
	} else {
		output.log.Errorf("Failed to convert interface{} to json.Number")
	}
	lsLinks[i] = lsLink
	keys[i] = lsLink.Key
	return i + 1, nil
}

func (output *ArangoOutput) updateLsLink(ctx context.Context, cursor driver.Cursor, command message.ArangoUpdateCommand) (error, []string, []arango.LSLink) {
	count := cursor.Count()
	keys := make([]string, count)
	lsLinks := make([]arango.LSLink, count)
	i := 0
	var err error
	for {
		i, err = output.processLsLinkDocument(ctx, cursor, command, lsLinks, keys, i)
		if driver.IsNoMoreDocuments(err) {
			return nil, keys[:i], lsLinks[:i]
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

func (output *ArangoOutput) updateDocuments(ctx context.Context, db driver.Database, command message.ArangoUpdateCommand, keys []string, lsLinks []arango.LSLink) error {
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

func (output *ArangoOutput) processArangoUpdateCommand(command message.ArangoUpdateCommand) {
	output.log.Infof("Processing Arango update for collection: %s", command.Collection)
	err, db := output.getDatabase()
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
		err, keys, lsLinks := output.updateLsLink(ctx, cursor, command)
		if err != nil {
			output.log.Errorf("Error updating ls_link: %v", err)
			return
		}
		output.log.Infof("Updating %d documents", len(keys))
		if err := output.updateDocuments(ctx, db, command, keys, lsLinks); err != nil {
			output.log.Errorf("Error updating documents: %v", err)
			return
		}
	} else {
		output.log.Errorf("Unknown collection: %s", command.Collection)
		return
	}
}

func (output *ArangoOutput) Start() {
	for command := range output.commandChan {
		switch cmdType := command.(type) {
		case message.ArangoUpdateCommand:
			output.processArangoUpdateCommand(cmdType)
		default:
			output.log.Errorf("Unknown command type: %v", command)
		}
	}
}

func (output *ArangoOutput) Stop() error {
	output.log.Debugln("Stopping ArangoDB output: ", output.config.Name)
	close(output.quitChan)
	return nil
}
