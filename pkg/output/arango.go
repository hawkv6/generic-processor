package output

import (
	"context"
	"fmt"
	"strings"

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

func (output *ArangoOutput) getAQLQuery(command message.ArangoUpdateCommand) string {
	var filters []string
	for key, val := range command.FilterBy {
		filters = append(filters, fmt.Sprintf("d.%s == '%s'", key, val))
	}
	filterStr := strings.Join(filters, " && ")

	var updateStr string
	if command.Index != nil {
		// If index is provided, update the specific index of the array
		updateStr = fmt.Sprintf("LET updatedArray = (FOR i IN RANGE(0, LENGTH(d.%s) - 1) RETURN i == %d ? %v : d.%s[i]) UPDATE d WITH { %s: updatedArray } IN %s", command.Field, *command.Index, command.Value, command.Field, command.Field, command.Collection)
	} else {
		// If index is not provided, update the whole field
		updateStr = fmt.Sprintf("UPDATE d WITH { %s: %v } IN %s", command.Field, command.Value, command.Collection)
	}

	query := fmt.Sprintf("FOR d IN %s FILTER %s %s", command.Collection, filterStr, updateStr)
	output.log.Infof("Generated AQL query: %s", query)
	return query
}

func (output *ArangoOutput) executeQuery(db driver.Database, query string) error {
	ctx := driver.WithWaitForSync(context.Background())
	_, err := db.Query(ctx, query, nil)
	if err != nil {
		output.log.Errorf("Error executing query: %v", err)
	}
	return err
}

func (output *ArangoOutput) processArangoUpdateCommand(command message.ArangoUpdateCommand) {
	output.log.Debugf("Processing ArangoUpdateCommand: %v", command)

	err, db := output.getDatabase()
	if err != nil {
		output.log.Errorf("Error getting database: %v", err)
		return
	}

	query := output.getAQLQuery(command)

	if err := output.executeQuery(db, query); err != nil {
		output.log.Errorf("Error executing query: %v", err)
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
