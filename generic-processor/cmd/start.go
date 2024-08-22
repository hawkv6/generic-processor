package cmd

import (
	"os"
	"os/signal"

	"github.com/hawkv6/generic-processor/pkg/config"
	"github.com/hawkv6/generic-processor/pkg/input"
	"github.com/hawkv6/generic-processor/pkg/output"
	"github.com/hawkv6/generic-processor/pkg/processor"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the generic processor",
	Run: func(cmd *cobra.Command, args []string) {
		config := config.NewDefaultConfig(configFile)
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

		inputManager.StartInputs()
		outputManager.StartOutputs()
		processorManager.StartProcessors()

		signalChan := make(chan os.Signal, 1)
		signal.Notify(signalChan, os.Interrupt)

		<-signalChan
		log.Info("Received interrupt signal, shutting down")
		inputManager.StopInputs()
		processorManager.StopProcessors()
		outputManager.StopOutputs()
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
	startCmd.Flags().StringVarP(&configFile, "config", "c", os.Getenv("HAWKV6_GENERIC_PROCESSOR_CONFIG"), "configuration file path e.g. config/example-config.yaml")
}
