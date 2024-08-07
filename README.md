<h1 align="center">containerlab telemetry linker</h1>
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


# generic-processor
The generic processor takes data from specified inputs, processes it, and sends it to specified outputs.

## Configuration

Used config on IOS-XR:
```
 sensor-group openconfig_interfaces
  sensor-path openconfig-interfaces:interfaces/interface/state/oper-status
  sensor-path openconfig-interfaces:interfaces/interface/subinterfaces/subinterface/openconfig-if-ip:ipv6/addresses/address
```

