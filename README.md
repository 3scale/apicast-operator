# apicast-operator

[![CircleCI](https://circleci.com/gh/3scale/apicast-operator/tree/master.svg?style=svg)](https://circleci.com/gh/3scale/apicast-operator/tree/master)
[![License](https://img.shields.io/badge/license-Apache--2.0-blue.svg)](http://www.apache.org/licenses/LICENSE-2.0)
[![GitHub release](https://img.shields.io/github/v/release/3scale/apicast-operator.svg)](https://github.com/3scale/apicast-operator/releases/latest)

## Overview

This project contains the APIcast operator software. APIcast operator is a piece of
software based on [Kubernetes operators](https://coreos.com/operators/) that
provides an easy way to install a 3scale APIcast self-managed solution, providing
configurability options at the time of installation.

The functionalities and definitions are provided via Kubernetes custom resources
which the operator is able to understand and process.

## Quickstart

To get up and running quickly, check our [Quickstart guides](doc/quickstart-guide.md)

## Features

Current capabilities state: Full Lifecycle

* Install a 3scale APIcast gateway self-managed solution
  * Using a 3scale [Porta](https://github.com/3scale/porta/) endpoint as
    the APIcast gateway configuration source
  * Providing the APIcast gateway configuration source from a configuration file
    via a pre-created Kubernetes Secret
* Upgrade from previously installed installed 3scale
  APIcast gateway self-managed solution
* Tune parameters of an already deployed 3scale APIcast gateway
  solution

## User Guide

Check our [Operator user guide](doc/operator-user-guide.md) for interacting with the APIcast operator

## Contributing

You can contribute by:

* Raising any issues you find using APIcast Operator
* Fixing issues by opening [Pull Requests](https://github.com/3scale/apicast-operator/pulls)
* Submitting a patch or opening a PR
* Improving documentation
* Talking about APIcast Operator

All bugs, tasks or enhancements are tracked as [GitHub issues](https://github.com/3scale/apicast-operator/issues).

The [Development guide](doc/development.md) describes how to build the APIcast Operator and how to test your changes before submitting a patch or opening a PR.


## Licensing

This software is licensed under the [Apache 2.0 license](https://www.apache.org/licenses/LICENSE-2.0).

See the LICENSE and NOTICE files that should have been provided along with this software for details.