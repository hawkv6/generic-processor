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

// func (output *ArangoOutput) getAQLQuery(command message.ArangoUpdateCommand) string {
// 	var filters []string
// 	for key, val := range command.FilterBy {
// 		filters = append(filters, fmt.Sprintf("d.%s == '%s'", key, val))
// 	}
// 	filterStr := strings.Join(filters, " && ")

// 	var updateStr string
// 	if command.Index != nil {
// 		updateStr = fmt.Sprintf("LET updatedArray = (FOR i IN RANGE(0, LENGTH(d.%s) - 1) RETURN i == %d ? %v : d.%s[i]) UPDATE d WITH { %s: updatedArray } IN %s", command.Field, *command.Index, command.Value, command.Field, command.Field, command.Collection)
// 	} else {
// 		updateStr = fmt.Sprintf("UPDATE d WITH { %s: %v } IN %s", command.Field, command.Value, command.Collection)
// 	}

// 	query := fmt.Sprintf("FOR d IN %s FILTER %s %s", command.Collection, filterStr, updateStr)
// 	output.log.Debugf("Generated AQL query: %s", query)
// 	return query
// }

// func (output *ArangoOutput) executeQuery(db driver.Database, query string) error {
// 	ctx := driver.WithWaitForSync(context.Background())
// 	_, err := db.Query(ctx, query, nil)
// 	if err != nil {
// 		output.log.Errorf("Error executing query: %v", err)
// 	}
// 	return err
// }

// func (output *ArangoOutput) processArangoUpdateCommand(command message.ArangoUpdateCommand) {
// 	output.log.Debugf("Processing Arango update for collection: %s", command.Collection)
// 	err, db := output.getDatabase()
// 	if err != nil {
// 		output.log.Errorf("Error getting database: %v", err)
// 		return
// 	}

// 	query := output.getAQLQuery(command)

// 	if err := output.executeQuery(db, query); err != nil {
// 		output.log.Errorf("Error executing query: %v", err)
// 	}
// }

func (output *ArangoOutput) updateLsLink(ctx context.Context, cursor driver.Cursor, command message.ArangoUpdateCommand) (error, []string, []arango.LSLink) {
	keys := make([]string, 0)
	lsLinks := make([]arango.LSLink, 0)
	counter := 0
	for {
		lsLink := arango.LSLink{}
		_, err := cursor.ReadDocument(ctx, &lsLink)
		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			output.log.Errorf("Error reading document: %v", err)
			return err, nil, nil
		}
		arangoUpdate, ok := command.Updates[lsLink.LocalLinkIP]
		if !ok {
			output.log.Errorf("No update found for local link IP: %s", lsLink.LocalLinkIP)
			continue
		}
		if arangoUpdate.Field == "unidir_link_delay" {
			if jsonNumber, ok := arangoUpdate.Value.(json.Number); ok {
				if intValue, err := jsonNumber.Int64(); err == nil {
					lsLink.UnidirLinkDelay = uint32(intValue)
				} else if floatValue, err := jsonNumber.Float64(); err == nil {
					lsLink.UnidirLinkDelay = uint32(floatValue)
				} else {
					output.log.Errorf("Failed to convert json.Number to int64 or float64")
				}
			} else {
				output.log.Errorf("Failed to convert interface{} to json.Number")
			}
		} else if arangoUpdate.Field == "unidir_link_delay_min_max" {
			if arangoUpdate.Index == nil || *arangoUpdate.Index > 1 {
				output.log.Errorf("Unknown index: %v", arangoUpdate.Index)
				continue
			} else if *arangoUpdate.Index == 0 {
				if jsonNumber, ok := arangoUpdate.Value.(json.Number); ok {
					if intValue, err := jsonNumber.Int64(); err == nil {
						lsLink.UnidirLinkDelayMinMax[0] = uint32(intValue)
					} else if floatValue, err := jsonNumber.Float64(); err == nil {
						lsLink.UnidirLinkDelayMinMax[0] = uint32(floatValue)
					} else {
						output.log.Errorf("Failed to convert json.Number to int64 or float64")
					}
				} else {
					output.log.Errorf("Failed to convert interface{} to json.Number")
				}
			} else if *arangoUpdate.Index == 1 {
				if jsonNumber, ok := arangoUpdate.Value.(json.Number); ok {
					if intValue, err := jsonNumber.Int64(); err == nil {
						lsLink.UnidirLinkDelayMinMax[1] = uint32(intValue)
					} else if floatValue, err := jsonNumber.Float64(); err == nil {
						lsLink.UnidirLinkDelayMinMax[1] = uint32(floatValue)
					} else {
						output.log.Errorf("Failed to convert json.Number to int64 or float64")
					}
				} else {
					output.log.Errorf("Failed to convert interface{} to json.Number")
				}
			}
		}
		lsLinks = append(lsLinks, lsLink)
		keys = append(keys, lsLink.Key)
		counter++
	}
	return nil, keys, lsLinks
}

func (output *ArangoOutput) processArangoUpdateCommand(command message.ArangoUpdateCommand) {
	output.log.Infof("Processing Arango update for collection: %s", command.Collection)
	// output.log.Debugf("Processing Arango update for collection: %s", command.Collection)
	err, db := output.getDatabase()
	if err != nil {
		output.log.Errorf("Error getting database: %v", err)
		return
	}
	ctx := context.Background()
	query := fmt.Sprintf("FOR d IN %s RETURN d", command.Collection)
	cursor, err := db.Query(ctx, query, nil)
	if err != nil {
		output.log.Errorf("Error executing query: %v", err)
		return
	}
	defer cursor.Close()

	if command.Collection == "ls_link" {
		err, keys, lsLinks := output.updateLsLink(ctx, cursor, command)
		if err != nil {
			output.log.Errorf("Error updating ls_link: %v", err)
			return
		}
		col, err := db.Collection(ctx, command.Collection)
		if err != nil {
			output.log.Errorf("Error getting collection: %v", err)
			return
		}
		_, errSlice, err := col.UpdateDocuments(ctx, keys, lsLinks)
		if err != nil {
			output.log.Errorf("Error updating documents: %v %v", errSlice, err)
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
