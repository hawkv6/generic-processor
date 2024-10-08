<h1 align="center">generic processor</h1>
<p align="center">
    <br>
    <img alt="GitHub Release" src="https://img.shields.io/github/v/release/hawkv6/generic-processor?display_name=release&style=flat-square">
    <img src="https://img.shields.io/badge/go%20report-A+-brightgreen.svg?style=flat-square">
    <img src="https://img.shields.io/github/actions/workflow/status/hawkv6/generic-processor/testing.yaml?style=flat-square&label=tests">
    <img src="https://img.shields.io/codecov/c/github/hawkv6/generic-processor?style=flat-square">
    <img src="https://img.shields.io/github/actions/workflow/status/hawkv6/generic-processor/golangci-lint.yaml?style=flat-square&label=checks">
</p>

<p align="center">
</p>

---

## Overview
[Jalapeno](https://github.com/cisco-open/jalapeno) utilizes various processors to manage network data. However, the standard processors are limited to handling BMP data for graph databases and streaming telemetry data to time series databases. Many use cases require more versatile processing capabilities, leading to the development of the generic processor.

The generic processor is designed to provide flexible processing abilities by allowing users to specify inputs (e.g., Kafka, ArangoDB, InfluxDB), process the data according to user-defined methods, and forward it to designated outputs (e.g., ArangoDB, Kafka, InfluxDB).

![Generic Processor Overview](docs/images/generic-processor-overview.drawio.svg)

## Design Considerations

The generic processor operates based on a validated configuration file, which initializes three core components:

- **input manager**: Manages data intake from specified sources.
- **processor manager**: Executes processing tasks as defined in the configuration.
- **output manager**: Handles the delivery of processed data to designated outputs.

Each component operates independently in separate Go routines, with communication primarily managed through channels. The generic processor is designed around a command-result pattern, where elements receive specific commands and return corresponding results. The configuration file also allows the same processor to be defined multiple times under different names, enabling parallel processing of the same data in various ways.

## Usage
```
generic-processor [command]
```
### Commands
- Validate the configuration file: [`validate`](/docs/commands/validate.md)
- Start the generic processor: [`start`](/docs/commands/start.md)
- Print the version `version`

## Available Processors
- [Telemetry to Arango Processor](docs/processors/telemetry-to-arango.md)


## Installation

### Using Package Manager
For Debian-based systems, install the package using apt:
```
sudo apt install ./generic_processor_{version}_amd64.deb
```

### Using Docker 
```
docker run --rm \
-v "hawkv6/generic-processor/config/:/config" \
-e HAWKV6_GENERIC_PROCESSOR_CONFIG=/config/example-config.yaml \ 
ghcr.io/hawkv6/generic-processor:latest start
```

### Using Binary
```
git clone https://github.com/hawkv6/generic-processor
cd generic-processor && make binary
./bin/generic-processor
```

## Getting Started

1. Deploy all necessary Kubernetes resources.
   - For more details, refer to the [hawkv6 deployment documentation](https://github.com/hawkv6/deployment).

2. Ensure the network is properly configured and operational.
   - Additional information can be found in the [hawkv6 testnetwork documentation](https://github.com/hawkv6/network).

3. Confirm that `clab-telemetry-linker` is active and running.
   - Detailed instructions are available in the [clab-telemetry-linker documentation](https://github.com/hawkv6/generic-processor).

4. Install the generic processor using one of the methods described above.

5. Launch the generic processor using the `start` command.


## Additional Information
- Environment variables are documented in the [`env documentation`](docs/env.md).
- An example configuration can be found in the [`config`](config) folder.
