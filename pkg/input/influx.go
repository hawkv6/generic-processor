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
	resultChan  chan message.ResultMessage
	quitChan    chan struct{}
	client      client.Client
}

func NewInfluxInput(config config.InfluxInputConfig, commandChan chan message.Command, resultChan chan message.ResultMessage) *InfluxInput {
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

func (input *InfluxInput) Start() {
	for {
		select {
		case <-input.quitChan:
			input.log.Infof("Stopping InfluxDB input '%s'", input.inputConfig.Name)
			return
		case <-input.commandChan:
			input.log.Debugln("Received command")
		}
	}
}

func (input *InfluxInput) Stop() error {
	close(input.quitChan)
	return nil
}
