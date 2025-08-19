# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [2.6.0](https://github.com/d0ugal/zigbee2mqtt-exporter/compare/v2.5.1...v2.6.0) (2025-08-19)


### Features

* **server:** add dynamic metrics information with collapsible interface ([2c4a1d4](https://github.com/d0ugal/zigbee2mqtt-exporter/commit/2c4a1d4946aa9826207ad27e774f5c136ef665d0))


### Bug Fixes

* **lint:** pre-allocate slices to resolve golangci-lint prealloc warnings ([cfd64ba](https://github.com/d0ugal/zigbee2mqtt-exporter/commit/cfd64bac7550aee59d04db8688cc7538013d838c))

## [2.5.1](https://github.com/d0ugal/zigbee2mqtt-exporter/compare/v2.5.0...v2.5.1) (2025-08-17)


### Bug Fixes

* resolve inconsistent label cardinality in DeviceInfo metric ([1193c94](https://github.com/d0ugal/zigbee2mqtt-exporter/commit/1193c9439b83d3a42893d900d5ee244b505dc4cf))

## [2.5.0](https://github.com/d0ugal/zigbee2mqtt-exporter/compare/v2.4.0...v2.5.0) (2025-08-17)


### Features

* add OTA update metrics and firmware version tracking ([7450ea0](https://github.com/d0ugal/zigbee2mqtt-exporter/commit/7450ea0e88bc2d6bdfb9fe203cca35e17251e544))

## [2.4.0](https://github.com/d0ugal/zigbee2mqtt-exporter/compare/v2.3.1...v2.4.0) (2025-08-17)


### Features

* improve reconnection logging and metrics ([df758c5](https://github.com/d0ugal/zigbee2mqtt-exporter/commit/df758c527c00f81077a689145a7ed4b11ad3ff57))
* improve reconnection logging formatting ([23ac456](https://github.com/d0ugal/zigbee2mqtt-exporter/commit/23ac456ecba0436f5e9be2e6e74bacb1a8b427ae))

## [2.3.1](https://github.com/d0ugal/zigbee2mqtt-exporter/compare/v2.3.0...v2.3.1) (2025-08-16)


### Bug Fixes

* apply golangci-lint formatting fixes ([506877d](https://github.com/d0ugal/zigbee2mqtt-exporter/commit/506877db4cb87a31a0e4d1fc7171c1e0be82e73b))

## [2.3.0](https://github.com/d0ugal/zigbee2mqtt-exporter/compare/v2.2.1...v2.3.0) (2025-08-16)


### Features

* add automerge rules for minor/patch updates and dev dependencies ([4123f22](https://github.com/d0ugal/zigbee2mqtt-exporter/commit/4123f224bd1ab1d5e7f44336c40fd0d817066753))
* upgrade to Go 1.25 ([cb09e56](https://github.com/d0ugal/zigbee2mqtt-exporter/commit/cb09e5643fb1653c2fe340c193d9410f8c9665e8))


### Bug Fixes

* revert golangci-lint config to version 2 for compatibility ([a66b3bb](https://github.com/d0ugal/zigbee2mqtt-exporter/commit/a66b3bb52925627745bf4b93c6ce04120f45b8de))
* update golangci-lint config for Go 1.25 compatibility ([a6cc61f](https://github.com/d0ugal/zigbee2mqtt-exporter/commit/a6cc61fd684789225ffe9a6366a3f1f649ea2a30))

## [2.2.1](https://github.com/d0ugal/zigbee2mqtt-exporter/compare/v2.2.0...v2.2.1) (2025-08-14)


### Bug Fixes

* ensure correct version reporting in release builds ([62e4312](https://github.com/d0ugal/zigbee2mqtt-exporter/commit/62e43129b5aea55bbc4d91f06abfe72e5885a07b))

## [2.2.0](https://github.com/d0ugal/zigbee2mqtt-exporter/compare/v2.1.0...v2.2.0) (2025-08-14)


### Features

* add version info metric and subtle version display in h1 header ([6700e8f](https://github.com/d0ugal/zigbee2mqtt-exporter/commit/6700e8f9eaa1d3644553ed179c011fc5fcbc4df8))
* add version to title, separate version info, and add copyright footer with GitHub links ([7df21f4](https://github.com/d0ugal/zigbee2mqtt-exporter/commit/7df21f40d9fc84b2e6c83c2dc3eb960c3cfd2f78))


### Bug Fixes

* update Dockerfile to inject version information and fix health endpoint to return JSON ([99538c0](https://github.com/d0ugal/zigbee2mqtt-exporter/commit/99538c07f5633f076a422114e233243bc539b96a))

## [2.1.0](https://github.com/d0ugal/zigbee2mqtt-exporter/compare/v2.0.0...v2.1.0) (2025-08-13)


### Features

* add battery level monitoring and alerts ([3cf3039](https://github.com/d0ugal/zigbee2mqtt-exporter/commit/3cf3039f0d5af4935c90090680da3251ed13f90a))
* add dynamic version information to web UI and CLI ([91815d8](https://github.com/d0ugal/zigbee2mqtt-exporter/commit/91815d8d2d19cf1b36e0502f4c90c1a5a258293c))

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
