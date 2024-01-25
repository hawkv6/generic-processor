package config

import "flag"

var configFlag string
var shortConfigFlag string

func init() {
	flag.StringVar(&configFlag, "config", "", "path to the config file e.g. /hawk-generic-processor/config.yaml")
	flag.StringVar(&shortConfigFlag, "c", "", "path to the config file (shorthand) e.g. /hawk-generic-processor/config.yaml")
}
