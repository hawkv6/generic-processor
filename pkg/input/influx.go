package input

import (
	"fmt"
	"time"

	"github.com/hawkv6/generic-processor/pkg/config"
	"github.com/hawkv6/generic-processor/pkg/logging"
	"github.com/hawkv6/generic-processor/pkg/message"
	"github.com/influxdata/influxdb/client/v2"
	"github.com/sirupsen/logrus"
)

type InfluxInput struct {
	log         *logrus.Entry
	inputConfig config.InfluxInputConfig
	commandChan chan message.Command
	resultChan  chan message.Result
	quitChan    chan struct{}
	client      client.Client
}

func NewInfluxInput(config config.InfluxInputConfig, commandChan chan message.Command, resultChan chan message.Result) *InfluxInput {
	return &InfluxInput{
		log:         logging.DefaultLogger.WithField("subsystem", Subsystem),
		inputConfig: config,
		commandChan: commandChan,
		resultChan:  resultChan,
		quitChan:    make(chan struct{}),
	}
}

func (input *InfluxInput) Init() error {
	client, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     input.inputConfig.URL,
		Username: input.inputConfig.Username,
		Password: input.inputConfig.Password,
		Timeout:  time.Duration(input.inputConfig.Timeout) * time.Second,
	})
	if err != nil {
		return err
	}

	_, _, err = client.Ping(0)
	if err != nil {
		return fmt.Errorf("error connecting to InfluxDB: %s", input.inputConfig.URL)
	}
	input.log.Debugf("Successfully created InfluxDB client '%s': ", input.inputConfig.Name)
	input.client = client
	return nil
}

func (input *InfluxInput) queryDB(query string) (res []client.Result, err error) {
	q := client.Query{
		Command:  query,
		Database: input.inputConfig.DB,
	}
	input.log.Debugln("Executing InfluxDB query: ", q)
	if response, err := input.client.Query(q); err == nil {
		if response.Error() != nil {
			return res, response.Error()
		}
		res = response.Results
	} else {
		return res, err
	}
	return res, nil
}

func (input *InfluxInput) createQuery(command message.InfluxQueryCommand) string {
	query := fmt.Sprintf(`SELECT %s("%s") FROM "%s" WHERE time > now() - %ds AND time <= now() GROUP BY`, command.Method, command.Field, command.Measurement, command.Interval)

	for i, name := range command.GroupBy {
		if i != 0 {
			query += ","
		}
		query += fmt.Sprintf(` "%s"`, name)
	}
	return query
}

func (input *InfluxInput) sendResults(results []client.Result, command message.InfluxQueryCommand) {
	input.log.Debugln("Sending InfluxDB results")
	resultMessage := message.InfluxResultMessage{
		OutputOptions: command.OutputOptions,
	}
	if len(results) == 0 {
		input.log.Errorf("No results returned from InfluxDB query")
		return
	}
	for _, row := range results[0].Series {
		for columnIndex, columnName := range row.Columns {
			if columnName == command.Method {
				if len(row.Values) == 0 {
					input.log.Errorf("No values returned from InfluxDB query")
					return
				}
				result := message.InfluxResult{
					Tags:  row.Tags,
					Value: row.Values[0][columnIndex],
				}
				resultMessage.Results = append(resultMessage.Results, result)
			}
		}
	}
	input.resultChan <- resultMessage
}

func (input *InfluxInput) executeCommand(command message.InfluxQueryCommand) {
	input.log.Debugf("Executing InfluxDB command: %v", command)
	query := input.createQuery(command)
	result, err := input.queryDB(query)
	if err != nil {
		input.log.Errorf("Error executing InfluxDB query: %v", err)
		return
	}
	if len(result) == 0 {
		input.log.Errorf("No results returned from InfluxDB query")
		return
	}
	if len(result[0].Series) == 0 {
		input.log.Errorf("No series returned from InfluxDB query")
		return
	}
	input.sendResults(result, command)
}

func (input *InfluxInput) Start() {
	for {
		select {
		case msg := <-input.commandChan:
			if influxCommand, ok := msg.(message.InfluxQueryCommand); ok {
				input.log.Debugf("Received InfluxCommand message: %v", influxCommand)
				input.executeCommand(influxCommand)
			} else {
				input.log.Errorf("Received invalid message type: %v", msg)
			}
		case <-input.quitChan:
			input.log.Infof("Stopping InfluxDB input '%s'", input.inputConfig.Name)
			return
		}
	}
}

func (input *InfluxInput) Stop() error {
	close(input.quitChan)
	return nil
}
