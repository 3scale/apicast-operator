# apicast-operator

[![CircleCI](https://circleci.com/gh/3scale/apicast-operator/tree/master.svg?style=svg)](https://circleci.com/gh/3scale/apicast-operator/tree/master)

This software is licensed under the [Apache 2.0 license](https://www.apache.org/licenses/LICENSE-2.0).

See the LICENSE and NOTICE files that should have been provided along with this software for details.

### Project Status: alpha

The project is currently alpha which means that there are still new features
and APIs planned that will be added in the future.
Due to this, breaking changes may still happen.

## Overview

This project contains the APIcast operator software. APIcast operator is a piece of
software based on [Kubernetes operators](https://coreos.com/operators/) that
provides:
* An easy way to install a 3scale APIcast self-managed solution, providing configurability
  options at the time of installation.

The functionalities and definitions are provided via Kubernetes custom resources
which the operator is able to understand and process.

## Prerequisites

* [operator-sdk] version v0.8.0.
* [git][git_tool]
* [go] version 1.12.5+
* [kubernetes] version v1.11.0+
* [kubectl] version v1.11+
* Access to a Kubernetes v1.11.0+ cluster.
* A user with administrative privileges in the Kubernetes cluster.

## Getting started

### Deploy APIcast gateway providing configuration secret

This is a basic example to deploy APIcast with a simple configuration.
The [json configuration file](https://github.com/3scale/APIcast/blob/master/examples/configuration/echo.json)
will make APIcast behave like an echo api endpoint.
See more [examples of configuration files](https://github.com/3scale/APIcast/tree/master/examples/configuration).

Create a secret with the configuration file:

```sh
curl https://raw.githubusercontent.com/3scale/APIcast/master/examples/configuration/echo.json -o $PWD/config.json
kubectl create secret generic apicast-echo-api-conf-secret --from-file=$PWD/config.json
```

Note that config file must be called `config.json`.
This is an [APIcast CRD reference](doc/apicast-crd-reference.md) requirement

Create an [APIcast custom resource](doc/apicast-crd-reference.md):

```yaml
$ cat my-echo-apicast.yaml
apiVersion: apps.3scale.net/v1alpha1
kind: APIcast
metadata:
  name: my-echo-apicast
spec:
  exposedHost:
    host: YOUR DOMAIN
  embeddedConfigurationSecretRef:
    name: apicast-echo-api-conf-secret

$ kubectl apply -f my-echo-apicast.yaml
```

### Deploy APIcast gateway providing 3scale portal URL

This is a basic example to deploy APIcast providing our 3scale Account Management API portal URL.

Create a secret with a URL. The URL format can be found [here](https://github.com/3scale/APIcast/blob/master/doc/parameters.md#threescale_portal_endpoint):

```sh
kubectl create secret generic apicast-portal-url --from-literal=AdminPortalURL=MY_3SCALE_URL
```

Note that secret key must be called `AdminPortalURL`.

Create an [APIcast custom resource](doc/apicast-crd-reference.md):

```yaml
$ cat my-supertest-apicast.yaml
apiVersion: apps.3scale.net/v1alpha1
kind: APIcast
metadata:
  name: my-supertest-apicast
spec:
  exposedHost:
    host: YOUR DOMAIN
  adminPortalCredentialsRef:
    name: apicast-portal-url

$ kubectl apply -f my-supertest-apicast.yaml
```

## Development

Run the operator locally with the default Kubernetes config file present at $HOME/.kube/config

Clone the project:

```sh
mkdir -p $GOPATH/src/github.com/3scale
cd $GOPATH/src/github.com/3scale
git clone https://github.com/3scale/apicast-operator.git
cd apicast-operator
```

An administrative user can create and deploy an APIcast CRD

```sh
kubectl create -f deploy/crds/apps_v1alpha1_apicast_crd.yaml
```

Create a new empty project

```sh
export NAMESPACE="operator-test"
$ cat operator-test.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: operator-test
  labels:
    name: operator-test

kubectl create -f operator-test.yaml
```

Create the ServiceAccount

```sh
kubectl create -f deploy/service_account.yaml
```

Create the roles and role bindings

```sh
kubectl create -f deploy/role.yaml
kubectl create -f deploy/role_binding.yaml
```

Install the dependencies

```sh
make vendor
```

```sh
make local
```

## Deploy nightly image to Kubernetes 1.13 using OLM
To install a nightly image (master tag) of this operator in Kubernetes 1.13 for end-to-end testing, 
create the [Operator Source](https://github.com/operator-framework/community-operators/blob/master/docs/testing-operators.md#linking-the-quay-application-repository-to-your-openshift-40-cluster)
provided in `deploy/olm-catalog/apicast-operatorsource.yaml` to load your operator bundle in Kubernetes.

```bash
kubectl create -f deploy/olm-catalog/apicast-operatorsource.yaml -n mynamespace
```

It will take a few minutes for the operator to become visible under the _OperatorHub_ section of the _Catalog_. It can be easily found by filtering the provider type to _Custom_.

## Documentation

* [APIcast CRD reference](doc/apicast-crd-reference.md)

[git_tool]:https://git-scm.com/downloads
[operator-sdk]:https://github.com/operator-framework/operator-sdk
[go]:https://golang.org/
[kubernetes]:https://kubernetes.io/
[kubectl]:https://kubernetes.io/docs/tasks/tools/install-kubectl/
