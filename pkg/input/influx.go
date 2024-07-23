package input

import (
	"encoding/json"
	"fmt"
	"sync"
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
	client      InfluxClient
	wg          sync.WaitGroup
}

func NewInfluxInput(config config.InfluxInputConfig, commandChan chan message.Command, resultChan chan message.Result) (*InfluxInput, error) {
	input := &InfluxInput{
		log:         logging.DefaultLogger.WithField("subsystem", Subsystem),
		inputConfig: config,
		commandChan: commandChan,
		resultChan:  resultChan,
		quitChan:    make(chan struct{}),
		wg:          sync.WaitGroup{},
	}
	client, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     input.inputConfig.URL,
		Username: input.inputConfig.Username,
		Password: input.inputConfig.Password,
		Timeout:  time.Duration(input.inputConfig.Timeout) * time.Second,
	})
	if err != nil {
		return nil, err
	}
	input.client = client
	return input, nil
}

func (input *InfluxInput) Init() error {
	_, _, err := input.client.Ping(time.Duration(input.inputConfig.Timeout))
	if err != nil {
		return fmt.Errorf("error connecting to InfluxDB: %s", input.inputConfig.URL)
	}
	input.log.Debugf("Successfully created InfluxDB client '%s': ", input.inputConfig.Name)
	return nil
}

func (input *InfluxInput) queryDB(query string) (res []client.Result, err error) {
	q := client.Query{
		Command:  query,
		Database: input.inputConfig.DB,
	}
	if response, err := input.client.Query(q); err == nil {
		if response.Error() != nil {
			return res, response.Error()
		}
		res = response.Results
		return res, nil
	} else {
		return res, err
	}
}

func (input *InfluxInput) createQuery(command message.InfluxQueryCommand) string {
	var query string
	if command.Transformation == nil {
		query = fmt.Sprintf(`SELECT %s("%s") FROM "%s" WHERE time > now() - %ds AND time <= now() GROUP BY`, command.Method, command.Field, command.Measurement, command.Interval)
	} else {
		query = fmt.Sprintf(`SELECT %s(%s("%s"),%ds) FROM "%s" WHERE time > now() - %ds AND time <= now() GROUP BY`, command.Transformation.Operation, command.Method, command.Field, command.Transformation.Period, command.Measurement, command.Interval)
	}

	for i, name := range command.GroupBy {
		if command.Transformation != nil && i == 0 {
			query += fmt.Sprintf(` time(%ds), `, command.Interval)
		}
		if i > 0 {
			query += ","
		}
		query += fmt.Sprintf(` "%s"`, name)
	}
	return query
}

func (input *InfluxInput) castValue(value interface{}) (float64, error) {
	jsonNumber, ok := value.(json.Number)
	if !ok {
		return 0, fmt.Errorf("value is not a json.Number")
	}
	return jsonNumber.Float64()
}

func (input *InfluxInput) getResults(results []client.Result, command message.InfluxQueryCommand, resultMessage *message.InfluxResultMessage) {
	for _, row := range results[0].Series {
		for columnIndex, columnName := range row.Columns {
			if columnName == command.Method || (command.Transformation != nil && command.Transformation.Operation == columnName) {
				if len(row.Values) == 0 {
					input.log.Errorf("No values returned from InfluxDB query")
					return
				}
				valueToConvert := row.Values[0][columnIndex]
				if value, err := input.castValue(valueToConvert); err != nil {
					input.log.Errorf("Failed to cast value: %v, error: %v", valueToConvert, err)
				} else {
					result := message.InfluxResult{
						Tags:  row.Tags,
						Value: value,
					}
					resultMessage.Results = append(resultMessage.Results, result)
				}
			}
		}
	}
}

func (input *InfluxInput) sendResults(results []client.Result, command message.InfluxQueryCommand) {
	resultMessage := message.InfluxResultMessage{
		OutputOptions: command.OutputOptions,
	}
	if len(results) == 0 {
		input.log.Errorf("No results returned from InfluxDB query")
		return
	}
	input.log.Infof("Retrieved %d results rows from InfluxDB query", len(results[0].Series))
	input.getResults(results, command, &resultMessage)
	input.resultChan <- resultMessage
}

func (input *InfluxInput) executeCommand(command message.InfluxQueryCommand) {
	query := input.createQuery(command)
	input.log.Debugf("Executing InfluxDB query: %s", query)
	result, err := input.queryDB(query)
	if err != nil {
		input.log.Errorf("Error executing InfluxDB query: %v", err)
		return
	}
	if len(result) == 0 || result[0].Series == nil {
		input.log.Errorf("No results or series returned from InfluxDB query")
		return
	}
	input.sendResults(result, command)
}

func (input *InfluxInput) Start() {
	input.log.Infof("Starting InfluxDB input '%s'", input.inputConfig.Name)
	input.wg.Add(1)
	for {
		select {
		case msg := <-input.commandChan:
			if influxCommand, ok := msg.(message.InfluxQueryCommand); ok {
				input.log.Debugln("Received InfluxCommand message")
				input.executeCommand(influxCommand)
			} else {
				input.log.Errorf("Received invalid message type: %v", msg)
			}
		case <-input.quitChan:
			input.log.Infof("Stopping InfluxDB input '%s'", input.inputConfig.Name)
			input.wg.Done()
			return
		}
	}
}

func (input *InfluxInput) Stop() error {
	close(input.quitChan)
	input.wg.Wait()
	if err := input.client.Close(); err != nil {
		return err
	}
	return nil
}
