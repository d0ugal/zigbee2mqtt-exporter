# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [2.0.0](https://github.com/d0ugal/zigbee2mqtt-exporter/compare/v1.0.1...v2.0.0) (2025-08-13)


### ⚠ BREAKING CHANGES

* **metrics:** Metric names updated to follow Prometheus best practices. This will break existing dashboards and alerting rules.

### Code Refactoring

* **metrics:** improve Prometheus metrics naming conventions ([01b02cb](https://github.com/d0ugal/zigbee2mqtt-exporter/commit/01b02cb84083b2124aa53cc0aac8c5d91df35195))

## [1.0.1](https://github.com/d0ugal/zigbee2mqtt-exporter/compare/v1.0.0...v1.0.1) (2025-08-13)


### Bug Fixes

* **docker:** update Alpine base image to 3.22.1 for better security and reproducibility ([75155a1](https://github.com/d0ugal/zigbee2mqtt-exporter/commit/75155a1c6ede1327c47495e868bfe760c5136140))

## 1.0.0 (2025-08-13)


### Bug Fixes

* resolve all linting issues for CI compliance ([225bd3b](https://github.com/d0ugal/zigbee2mqtt-exporter/commit/225bd3b2a326eddb3e780435d19fb2c74408d7c4))

## [Unreleased]

### Added
- Initial release of zigbee2mqtt-exporter
- Prometheus metrics export for Zigbee2MQTT devices
- WebSocket connection to Zigbee2MQTT
- Configuration via environment variables
- Docker support

### Changed

### Deprecated

### Removed

### Fixed

### Security
