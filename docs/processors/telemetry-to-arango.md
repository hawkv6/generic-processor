# Telemetry to Arango Processor

## Overview
The Telemetry to Arango Processor is designed to enrich BMP data stored in the graph database with additional data received via streaming telemetry (MDT/YANG-PUSH).

![Telemetry To Arango Processor Overview](../images/telemetry-to-arango-processor-overview.drawio.svg)

## Functionality

The telemetry to Arango processor consists of several key components:

- **Kafka Consumer**: Consumes interface data from Kafka in an event-driven manner and forwards it to the telemetry to Arango processor. This allows it to respond to changes in interface status or IPv6 addresses.
  
- **Influx Client**: Periodically consumes time series data from InfluxDB, with the period defined by the `interval` setting in the configuration file.
  
- **Arango Client**: Updates links in the graph database with the specific calculated values.

- **Kafka Producer**: Produces and sends messages about updated links to inform other systems of the changes. Additionally, it sends normalization data to Kafka, which is then forwarded to InfluxDB.

- **Normalization Process**: The processor applies IQR-based min-max normalization to the telemetry data. This process involves the following steps:
  - **Interquartile Range Calculation**: The processor calculates the interquartile range (IQR) of the data, which is the range between the 25th percentile (Q1) and the 75th percentile (Q3).
  - **Min-Max Adjustment**: Using the IQR, the processor adjusts the minimum and maximum values for the data fields. This adjustment helps in reducing the impact of outliers by focusing on the more central data points.
  - **Data Normalization**: The adjusted min and max values are used to normalize the data fields, ensuring that the data remains within a defined range and is less sensitive to extreme values.

  The normalized data is then updated in the ArangoDB graph database and sent as normalized metrics to Kafka. This ensures that the data is consistently formatted and ready for accurate analysis.

These components work together to enrich links with data from InfluxDB, enhancing the overall dataset and ensuring that the data is robustly normalized, making it more reliable for downstream processes.


## Prerequisites
Ensure you have the Kafka, Influx and Kafka up and running. For further documentation look at the [deployment guide](https://github.com/hawkv6/deployment) to install and configure the K8S resources.


## Configuration

The following configuration parameters are required for the telemetry to Arango processor.

### Input Configuration

#### Kafka Configuration
- **\<kafka-input-name>**:
  - **broker**: Specify the Kafka broker address (e.g., `broker-hostname:port`).
  - **topic**: The name of the Kafka topic to consume interface data from.

This Kafka input configuration is used to consume interface data from the specified Kafka topic.

#### InfluxDB Configuration
- **\<influxdb-input-name>**:
  - **url**: The URL of the InfluxDB instance (e.g., `http://influxdb-hostname:port`).
  - **db**: The name of the database in InfluxDB to query from.
  - **username**: The username to authenticate with InfluxDB.
  - **password**: The password associated with the username.

This InfluxDB input configuration is used to periodically consume time series data from the specified InfluxDB database.

### Output Configuration

#### ArangoDB Output
- **\<arango-output-name>**:
  - **url**: The URL of the ArangoDB instance (e.g., `https://arangodb-hostname/`).
  - **database**: The name of the database in ArangoDB to update.
  - **username**: The username to authenticate with ArangoDB.
  - **password**: The password associated with the username.
  - **skip_tls_verification**: A boolean value (`true` or `false`) to specify whether to skip TLS verification.

This configuration defines how processed data is stored in the specified ArangoDB instance.

#### Kafka Output Configuration
- **\<kafka-output-event-name>**:
  - **broker**: Specify the Kafka broker address (e.g., `broker-hostname:port`).
  - **topic**: The name of the Kafka topic to send event notifications to.
  - **type**: The type of Kafka output, typically `event-notification`.

This Kafka output configuration is used to send link update events.

- **\<kafka-output-normalization-name>**:
  - **broker**: Specify the Kafka broker address (e.g., `broker-hostname:port`).
  - **topic**: The name of the Kafka topic to send normalized telemetry metrics to.
  - **type**: The type of Kafka output, typically `normalization`.

This Kafka output configuration is used to send normalized telemetry metrics.

### Processor Configuration

#### TelemetryToArango Processor
- **\<processor-name>**:
  - **inputs**:
    - List the names of the input configurations defined above (e.g., `<kafka-input-name>`, `<influxdb-input-name>`).
  - **outputs**:
    - List the names of the output configurations defined above (e.g., `<arango-output-name>`, `<kafka-output-event-name>`, `<kafka-output-normalization-name>`).
  - **interval**: Define the processing interval in seconds (e.g., `30` seconds).
  - **normalization**:
    - **field-mappings**:
      - Specify the mapping between input fields and normalized fields (e.g., `input-field-name`: `normalized-field-name`).
  - **modes**:
    - **input-options**:
      - **<influxdb-input-name>**:
        - **measurement**: Specify the InfluxDB measurement to query from.
        - **field**: The field within the measurement to be processed.
        - **method**: The aggregation method to apply (e.g., `median`, `mean`, `max`, `min`, `last`).
        - **group_by**: List the fields to group the data by (e.g., `interface_name`, `source`).
    - **output-options**:
      - **<arango-output-name>**:
        - **collection**: The name of the ArangoDB collection to update.
        - **method**: The method to use for updating the collection (e.g., `update`).
        - **filter_by**: List the fields to filter the data by before updating (e.g., `local_link_ip`).
        - **field**: The field in the ArangoDB collection to update with the processed data.

This processor configuration defines how telemetry data is processed, including mapping fields from InfluxDB measurements to corresponding fields in ArangoDB collections. An example configuration can be found in the `config` folder.

### YANG Models

The following YANG model configuration is used on IOS-XR:

```bash
sensor-group openconfig_interfaces
  sensor-path openconfig-interfaces:interfaces/interface/state/oper-status
  sensor-path openconfig-interfaces:interfaces/interface/subinterfaces/subinterface/openconfig-if-ip:ipv6/addresses/address
```
