# Validate 

## Overview
The `validate` command validates the provided configuration file.

## Command Syntax
```bash
generic-processor validate -c <config file>
```

`--config <config file>` or `-c <config file>`: Specifies the configuration file location if not set via the `HAWKV6_GENERIC_PROCESSOR_CONFIG` environment variable.


## Example 
```
generic-processor validate -c example-config.yaml
INFO[2024-08-09T11:14:55Z] Validating influx input 'jalapeno-influx'     subsystem=config
INFO[2024-08-09T11:14:55Z] Successfully validated influx input 'jalapeno-influx'  subsystem=config
INFO[2024-08-09T11:14:55Z] Successfully validated all influx inputs      subsystem=config
INFO[2024-08-09T11:14:55Z] Validating kafka input 'jalapeno-kafka-openconfig'  subsystem=config
INFO[2024-08-09T11:14:55Z] Successfully validated kafka input 'jalapeno-kafka-openconfig'  subsystem=config
INFO[2024-08-09T11:14:55Z] Successfully validated all kafka inputs       subsystem=config
INFO[2024-08-09T11:14:55Z] Successfully validated all inputs             subsystem=config
INFO[2024-08-09T11:14:55Z] Validating arango output 'jalapeno-arango'    subsystem=config
INFO[2024-08-09T11:14:55Z] Successfully validated arango output 'jalapeno-arango'  subsystem=config
INFO[2024-08-09T11:14:55Z] Successfully validated all arango outputs     subsystem=config
INFO[2024-08-09T11:14:55Z] Validating kafka output 'jalapeno-kafka-lslink-events'  subsystem=config
INFO[2024-08-09T11:14:55Z] Successfully validated kafka output 'jalapeno-kafka-lslink-events'  subsystem=config
INFO[2024-08-09T11:14:55Z] Validating kafka output 'jalapeno-kafka-normalization-metrics'  subsystem=config
INFO[2024-08-09T11:14:55Z] Successfully validated kafka output 'jalapeno-kafka-normalization-metrics'  subsystem=config
INFO[2024-08-09T11:14:55Z] Successfully validated all kafka outputs      subsystem=config
INFO[2024-08-09T11:14:55Z] Successfully validated all outputs            subsystem=config
INFO[2024-08-09T11:14:55Z] Successfully validated TelemetryToArango processor 'LsLink'  subsystem=config
INFO[2024-08-09T11:14:55Z] Successfully validated all TelemetryToArango processors  subsystem=config
INFO[2024-08-09T11:14:55Z] Successfully validated all processors         subsystem=config
INFO[2024-08-09T11:14:55Z] Config file successfully validated            subsystem=main
```