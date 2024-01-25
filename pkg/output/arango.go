package output

import (
	"fmt"

	"github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/http"
	"github.com/hawkv6/generic-processor/pkg/config"
	"github.com/hawkv6/generic-processor/pkg/logging"
	"github.com/sirupsen/logrus"
)

type ArangoOutput struct {
	log    *logrus.Entry
	config config.ArangoOutputConfig
	client driver.Client
}

func NewArangoOutput(config config.ArangoOutputConfig) *ArangoOutput {
	return &ArangoOutput{
		log:    logging.DefaultLogger.WithField("subsystem", Subsystem),
		config: config,
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

func (output *ArangoOutput) Start() {
}

func (output *ArangoOutput) Stop() {
}
