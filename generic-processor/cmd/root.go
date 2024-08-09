package cmd

import (
	"github.com/hawkv6/generic-processor/pkg/logging"
	"github.com/spf13/cobra"
)

var log = logging.DefaultLogger.WithField("subsystem", "main")
var configFile string

var rootCmd = &cobra.Command{
	Use:   "generic-processor",
	Short: "Generic processor takes data from specified inputs, processes it, and sends it to specified outputs.",
	Long: `
Generic processor takes data from specified inputs, processes it, and sends it to specified outputs.

Validate the config file
$ hawkeye validate -c <path-to-config-file>

Start the generic processor 
$ hawkeye start -c <path-to-config-file>
`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
