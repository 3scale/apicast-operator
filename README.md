# apicast-operator

[![CircleCI](https://circleci.com/gh/3scale/apicast-operator/tree/master.svg?style=svg)](https://circleci.com/gh/3scale/apicast-operator/tree/master)

This software is licensed under the [Apache 2.0 license](https://www.apache.org/licenses/LICENSE-2.0).

See the LICENSE and NOTICE files that should have been provided along with this software for details.

### Project Status: alpha

The project is currently alpha which means that there are still new feautres
and APIs planned that will be added in the future.
Due to this, breaking changes may still happen.

Only use for short-lived testing clusters. Do not deploy it in the same
OpenShift project than one having an already existing
3scale APIcast solution as it could potentially alter/delete the
existing elements in the project.

## Overview

This project contains the APIcast operator software. APIcast operator is a
software based on [Kubernetes operators](https://coreos.com/operators/) that
provides:
* A way to install a 3scale APIcast self-managed solution, providing configurability
  options at the time of installation

This functionalities definitions are provided via Kubernetes custom resources
that the operator is able to understand and process.

## Prerequisites

* [operator-sdk] version v0.8.0.
* [git][git_tool]
* [go] version 1.12.5+
* [kubernetes] version v1.11.0+
* [oc] version v3.11+
* Access to a Openshift v3.11.0+ cluster.
* A user with administrative privileges in the OpenShift cluster.

## Getting started

WORK IN PROGRESS
