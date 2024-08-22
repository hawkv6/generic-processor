package cmd

import (
	"os"

	"github.com/hawkv6/generic-processor/pkg/config"
	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validates the provided config file",
	Run: func(cmd *cobra.Command, args []string) {
		config := config.NewDefaultConfig(configFile)
		if err := config.Read(); err != nil {
			log.Fatalf("error reading config: %v", err)
		}
		if err := config.Validate(); err != nil {
			log.Fatalf("error validating config: %v", err)
		}
		log.Infoln("Config file successfully validated")
	},
}

func init() {
	rootCmd.AddCommand(validateCmd)
	validateCmd.Flags().StringVarP(&configFile, "config", "c", os.Getenv("HAWKV6_GENERIC_PROCESSOR_CONFIG"), "configuration file path e.g. config/example-config.yaml")
}
