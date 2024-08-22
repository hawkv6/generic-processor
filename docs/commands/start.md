# Start

## Overview
The `start` command initiates all input, processor, and output operations as defined in the configuration file.

## Command Syntax
```bash
generic-processor start -c <config file>
```

`--config <config file>` or `-c <config file>`: Specifies the configuration file location if not set via the `HAWKV6_GENERIC_PROCESSOR_CONFIG` environment variable.

## Example 
```
ins@dnaserver9:~/hawkv6/generic-processor/config$ generic-processor start -c example-config.yaml
INFO[2024-08-09T11:16:33Z] Validating influx input 'jalapeno-influx'     subsystem=config
INFO[2024-08-09T11:16:33Z] Successfully validated influx input 'jalapeno-influx'  subsystem=config
INFO[2024-08-09T11:16:33Z] Successfully validated all influx inputs      subsystem=config
INFO[2024-08-09T11:16:33Z] Validating kafka input 'jalapeno-kafka-openconfig'  subsystem=config
INFO[2024-08-09T11:16:33Z] Successfully validated kafka input 'jalapeno-kafka-openconfig'  subsystem=config
INFO[2024-08-09T11:16:33Z] Successfully validated all kafka inputs       subsystem=config
INFO[2024-08-09T11:16:33Z] Successfully validated all inputs             subsystem=config
INFO[2024-08-09T11:16:33Z] Validating arango output 'jalapeno-arango'    subsystem=config
INFO[2024-08-09T11:16:33Z] Successfully validated arango output 'jalapeno-arango'  subsystem=config
INFO[2024-08-09T11:16:33Z] Successfully validated all arango outputs     subsystem=config
INFO[2024-08-09T11:16:33Z] Validating kafka output 'jalapeno-kafka-normalization-metrics'  subsystem=config
INFO[2024-08-09T11:16:33Z] Successfully validated kafka output 'jalapeno-kafka-normalization-metrics'  subsystem=config
INFO[2024-08-09T11:16:33Z] Validating kafka output 'jalapeno-kafka-lslink-events'  subsystem=config
INFO[2024-08-09T11:16:33Z] Successfully validated kafka output 'jalapeno-kafka-lslink-events'  subsystem=config
INFO[2024-08-09T11:16:33Z] Successfully validated all kafka outputs      subsystem=config
INFO[2024-08-09T11:16:33Z] Successfully validated all outputs            subsystem=config
INFO[2024-08-09T11:16:33Z] Successfully validated TelemetryToArango processor 'LsLink'  subsystem=config
INFO[2024-08-09T11:16:33Z] Successfully validated all TelemetryToArango processors  subsystem=config
INFO[2024-08-09T11:16:33Z] Successfully validated all processors         subsystem=config
INFO[2024-08-09T11:16:33Z] Starting all inputs                           subsystem=inputs
INFO[2024-08-09T11:16:33Z] Starting all outputs                          subsystem=output
INFO[2024-08-09T11:16:33Z] Starting all processors                       subsystem=processors
INFO[2024-08-09T11:16:33Z] Starting Kafka input 'jalapeno-kafka-openconfig'  subsystem=inputs
INFO[2024-08-09T11:16:33Z] Starting InfluxDB input 'jalapeno-influx'     subsystem=inputs
INFO[2024-08-09T11:16:33Z] Starting TelemetryToArangoProcessor 'LsLink'  subsystem=processors
INFO[2024-08-09T11:16:33Z] Starting scheduling Influx commands for 'jalapeno-influx' input every 30 seconds  subsystem=processors
...
^CINFO[2024-08-09T11:16:37Z] Received interrupt signal, shutting down      subsystem=main
INFO[2024-08-09T11:16:37Z] Stopping all inputs                           subsystem=inputs
INFO[2024-08-09T11:16:37Z] Stopping InfluxDB input 'jalapeno-influx'     subsystem=inputs
INFO[2024-08-09T11:16:37Z] Stopping Kafka input 'jalapeno-kafka-openconfig'  subsystem=inputs
INFO[2024-08-09T11:16:37Z] All inputs stopped                            subsystem=inputs
INFO[2024-08-09T11:16:37Z] Stopping processors                           subsystem=processors
INFO[2024-08-09T11:16:37Z] Stopping TelemetryToArangoProcessor 'LsLink'  subsystem=processors
INFO[2024-08-09T11:16:37Z] Stopping Kafka OpenConfig Processor           subsystem=processors
INFO[2024-08-09T11:16:37Z] All processors stopped                        subsystem=processors
INFO[2024-08-09T11:16:37Z] Stopping all outputs                          subsystem=output
INFO[2024-08-09T11:16:37Z] Stopping processing 'jalapeno-kafka-openconfig' input messages  subsystem=processors
INFO[2024-08-09T11:16:37Z] Stopping Processing 'jalapeno-influx' input messages  subsystem=processors
INFO[2024-08-09T11:16:37Z] Stopping scheduling Influx commands for 'jalapeno-influx' input  subsystem=processors
INFO[2024-08-09T11:16:37Z] All output stopped      
```