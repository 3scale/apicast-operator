# Change Log
All notable changes to this project will be documented in this file.
This project adheres to [Semantic Versioning](http://semver.org/).

## [Unreleased]

## [0.9.0] - 2024-03-12

### Added

* Horizontal Pod Autoscaling support [#207](https://github.com/3scale/apicast-operator/pull/207)

## [0.7.0] - 2022-10-11

### Fixed

* Enhance replica reconciliation [#178](https://github.com/3scale/apicast-operator/pull/178)

## [0.6.0] - 2022-08-23

### Added

* Metering labels [#155](https://github.com/3scale/apicast-operator/pull/155)
* Smart secret management [#170](https://github.com/3scale/apicast-operator/pull/170)
* Cluster scope deployment [#174](https://github.com/3scale/apicast-operator/pull/174)

## [0.5.1] - 2021-12-13

### Added

* New configuration param Apicast CRD: `APICAST_UPSTREAM_RETRY_CASES` [#92](https://github.com/3scale/apicast-operator/pull/92)
* New configuration param Apicast CRD: `APICAST_CACHE_MAX_TIME` [#92](https://github.com/3scale/apicast-operator/pull/92)
* New configuration param Apicast CRD: `APICAST_CACHE_STATUS_CODES` [#92](https://github.com/3scale/apicast-operator/pull/92)
* New configuration param Apicast CRD: `APICAST_OIDC_LOG_LEVEL` [#92](https://github.com/3scale/apicast-operator/pull/92)
* New configuration param Apicast CRD: `APICAST_LOAD_SERVICES_WHEN_NEEDED` [#92](https://github.com/3scale/apicast-operator/pull/92)
* New configuration param Apicast CRD: `APICAST_SERVICES_FILTER_BY_URL` [#92](https://github.com/3scale/apicast-operator/pull/92)
* New configuration param Apicast CRD: `APICAST_SERVICE_${ID}_CONFIGURATION_VERSION` [#92](https://github.com/3scale/apicast-operator/pull/92)
* Enable TLS at pod level [#96](https://github.com/3scale/apicast-operator/pull/96)
* Add 'app' label to the APIcast CRD [#102](https://github.com/3scale/apicast-operator/pull/102)
* Enable workers setting [#106](https://github.com/3scale/apicast-operator/pull/106)
* Add default zap logger flags as part of the operator command flags [#110](https://github.com/3scale/apicast-operator/pull/110)
* Add timezone attribute to control APIcast pods local timezone [#111](https://github.com/3scale/apicast-operator/pull/111)
* Custom policies [#126](https://github.com/3scale/apicast-operator/pull/126)
* Custom environments [#130](https://github.com/3scale/apicast-operator/pull/130)
* Extended metrics [#132](https://github.com/3scale/apicast-operator/pull/132)
* Add Opentracing configurability [#141](https://github.com/3scale/apicast-operator/pull/141) [#156](https://github.com/3scale/apicast-operator/pull/156)
* Add proxy-related attributes to APIcast CR [#159](https://github.com/3scale/apicast-operator/pull/159)
* bundle common label: `app=apicast` [#164](https://github.com/3scale/apicast-operator/pull/164)

### Changed

* Operator-sdk upgrade to 1.1 [#88](https://github.com/3scale/apicast-operator/pull/88)
* Operator-sdk upgrade to 1.2 [#99](https://github.com/3scale/apicast-operator/pull/99)
* Update CustomResourceDefinition API version to v1 [#104](https://github.com/3scale/apicast-operator/pull/104)
* Ingress to networking v1 [#148](https://github.com/3scale/apicast-operator/pull/148)
* k8s libraries v0.19.14 [#148](https://github.com/3scale/apicast-operator/pull/148)
* Delete kube-rbac-proxy container from controller-manager [#165](https://github.com/3scale/apicast-operator/pull/165)

### Fixed

* Fix status reconciliation when deployment does not exist [#97](https://github.com/3scale/apicast-operator/pull/97)
* CVE-2020-14040 [#113](https://github.com/3scale/apicast-operator/pull/113)
* CVE-2020-9283 [#115](https://github.com/3scale/apicast-operator/pull/115)
* Fix reconcile service https port [#142](https://github.com/3scale/apicast-operator/pull/142)

## [0.4.0] - 2021-05-05

### Added
- APIcast v3.10 [#120](https://github.com/3scale/apicast-operator/pull/120)
- Apicast workers [#119](https://github.com/3scale/apicast-operator/pull/119)
- Resource requirements [#79](https://github.com/3scale/apicast-operator/pull/79)

## [0.3.0] - 2020-11-02

### Added
- APIcast v3.9 [#89](https://github.com/3scale/apicast-operator/pull/89)

### Changed
- Operator SDK v0.15.2 [#54](https://github.com/3scale/apicast-operator/pull/54)
- Go version v1.13 [#54](https://github.com/3scale/apicast-operator/pull/54)
- CSV: describe additional functionalities [#73](https://github.com/3scale/apicast-operator/pull/73)

### Fixed
- Fix Ingress deletion when exposedHost is deleted from CR [#69](https://github.com/3scale/apicast-operator/pull/69)

## [0.2.0] - 2020-04-02

### Added
- Add image field to APIcast status [#41](https://github.com/3scale/apicast-operator/pull/41)
- APIcast v3.8 [#70](https://github.com/3scale/apicast-operator/pull/70)
- Ingress support OCP [#70](https://github.com/3scale/apicast-operator/pull/70)
- Go v1.12 [#70](https://github.com/3scale/apicast-operator/pull/70)

## [0.1.0] - 2019-11-02

### Added
- APIcast v3.7

[Unreleased]: https://github.com/3scale/apicast-operator/compare/v0.7.0...HEAD
[0.7.0]: https://github.com/3scale/apicast-operator/releases/tag/v0.7.0
[0.6.0]: https://github.com/3scale/apicast-operator/releases/tag/v0.6.0
[0.5.1]: https://github.com/3scale/apicast-operator/releases/tag/v0.5.1
[0.4.0]: https://github.com/3scale/apicast-operator/releases/tag/v0.4.0
[0.3.0]: https://github.com/3scale/apicast-operator/releases/tag/v0.3.0
[0.2.0]: https://github.com/3scale/apicast-operator/releases/tag/v0.2.0
[0.1.0]: https://github.com/3scale/apicast-operator/releases/tag/v0.1.0
