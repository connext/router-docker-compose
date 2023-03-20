# General info

This is development/beta version of Prometheus exporter for Connext Network Routers.

# Building, configuring and running exporter

## Building

Exporter can be built from source `connext-exporter.go` file or used as a precompilled `exporter` binary

## Configuration

Exporter uses 2 config files: one for exporter itself and one for the Router (grabbing RPC endpoints from there)

Exporter config file has the following structure
```
{
  "network": "mainnet",
	"Router": "0x97bxxxx",
	"RouterScrapeInterval": 60,
	"RPCScrapeInterval": 300,
	"ETHScrapeInterval": 3600,
	"RouterQueryLimit": 100,
	"NetworkQueryLimit": 1000,
	"Host": "localhost",
	"Port": 9999
}
```

Configuration options explanation:

* `network`:  Router network, can be `mainnet` or `testnet`
* `Router`: Router address like `0xf26c772c0ff3a6036bddabdaba22cf65eca9f97c`
* `RouterScrapeInterval`: interval for scraping Router's metrics in seconds
* `RPCScrapeInterval`: interval for scraping RPC's metrics in seconds
* `ETHScrapeInterval`: interval for scraping ETH price in seconds
* `RouterQueryLimit`: limit for Router stats query
* `NetworkQueryLimit`: limit for Connext stats query
* `Host`: host for exporter to bind to
* `Port`: port for exporter to bind to

## Running exporter

Exporter can be run from CLI like:

```
exporter -c config.json -r ../router/config.json
```

alternatively you can use `systemd.config` file or build docker container.

# Prometheus and Grafana configuration

In order to use all the power of this exporter, connect is to your Prometheus instance as follows

```
- targets:
  - 'connext_router_address:9999'  #exporter address and port
  labels:
    server: exporter
```

You can connect multiple exporters configured for multiple routers this way.

### Alerts

Example of Prometheus alerts can be found in `alert.rules` file

### Grafana dashboard

Grafana dashboard can be found in `Router.json` file
