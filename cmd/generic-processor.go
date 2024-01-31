package main

import (
	"os"
	"os/signal"

	"github.com/hawkv6/generic-processor/pkg/config"
	"github.com/hawkv6/generic-processor/pkg/input"
	"github.com/hawkv6/generic-processor/pkg/logging"
	"github.com/hawkv6/generic-processor/pkg/output"
	"github.com/hawkv6/generic-processor/pkg/processor"
)

var log = logging.DefaultLogger.WithField("subsystem", "cmd")

func main() {
	config := config.NewDefaultConfig()
	if err := config.Read(); err != nil {
		log.Fatalf("error reading config: %v", err)
	}
	if err := config.Validate(); err != nil {
		log.Fatalf("error validating config: %v", err)
	}

	inputManager := input.NewDefaultInputManager(config)
	if err := inputManager.InitInputs(); err != nil {
		log.Fatalf("error initializing inputs: %v", err)
	}
	outputManager := output.NewDefaultOutputManager(config)
	if err := outputManager.InitOutputs(); err != nil {
		log.Fatalf("error initializing outputs: %v", err)
	}

	processorManager := processor.NewDefaultProcessorManager(config, inputManager, outputManager)
	if err := processorManager.Init(); err != nil {
		log.Fatalf("error initializing processors: %v", err)
	}

	if err := inputManager.StartInputs(); err != nil {
		log.Fatalf("error starting inputs: %v", err)
	}

	if err := outputManager.StartOutputs(); err != nil {
		log.Fatalf("error starting outputs: %v", err)
	}

	processorManager.StartProcessors()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	<-signalChan
	log.Info("Received interrupt signal, shutting down")
	inputManager.StopInputs()
	processorManager.StopProcessors()

}
