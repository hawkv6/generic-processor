# Telemetry To Arango Processor
## Configuration

Used config on IOS-XR:
```
 sensor-group openconfig_interfaces
  sensor-path openconfig-interfaces:interfaces/interface/state/oper-status
  sensor-path openconfig-interfaces:interfaces/interface/subinterfaces/subinterface/openconfig-if-ip:ipv6/addresses/address
```