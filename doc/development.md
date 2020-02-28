# Development Guide

## Table of contents
* [Prerequisites](#Prerequisites)
* [Clone repository](#Clone-repository)
* [Building APIcast operator image](#building-APIcast-operator-image)
* [Run APIcast Operator](#run-APIcast-operator)
  * [Run APIcast Operator Locally](#run-APIcast-operator-locally)
  * [Deploy and run APIcast Operator Manually](#deploy-and-run-APIcast-operator-manually)
    * [Cleanup manually deployed operator](#cleanup-manually-deployed-operator)
  * [Deploy custom APIcast Operator using OLM](#deploy-custom-APIcast-operator-using-olm)
* [Run tests](#run-tests)
* [Manifest management](#manifest-management)
  * [Verify operator manifest](#verify-operator-manifest)
  * [Push an operator bundle into external app registry](#push-an-operator-bundle-into-external-app-registry)
* [Licenses management](#licenses-management)
  * [Manually adding a new license](#manually-adding-a-new-license)

## Prerequisites

* [operator-sdk] version v0.15.2
* [git][git_tool]
* [go] version 1.13+
* [kubernetes] version v1.11.0+
* [kubectl] version v1.11+
* Access to a Kubernetes v1.11.0+ cluster.
* A user with administrative privileges in the Kubernetes cluster.

## Clone repository

```sh
mkdir -p $GOPATH/src/github.com/3scale
cd $GOPATH/src/github.com/3scale
git clone https://github.com/3scale/apicast-operator
cd apicast-operator
git checkout master
```
## Building APIcast operator image

[Clone the repo](#Clone-repository)

Build operator image

```sh
make build IMAGE=quay.io/myorg/apicast-operator VERSION=test
```

## Run APIcast Operator

### Run APIcast Operator Locally

Run the operator locally with the default Kubernetes config file present at $HOME/.kube/config

Run operator from command line, it will not be deployed as pod.

* [Clone the repo](#Clone-repository)

* Register the APIcast operator CRD in the Kubernetes API Server

```sh
// As a cluster admin
for i in `ls deploy/crds/*_crd.yaml`; do kubectl create -f $i ; done
```

* Create a new Kubernetes namespace (optional)

```sh
export NAMESPACE=operator-test
kubectl create namespace ${NAMESPACE}

Do not forget to change the current Kubernetes context to the newly
created namespace or create a new context for it and set the new one as the
active context
```

* Install the dependencies

```sh
make vendor
```

* Run operator

```sh
make local
```

### Deploy and run APIcast Operator Manually

Build operator image and deploy manually as a pod

* [Build APIcast operator image](#Building-APIcast-operator-image)

* Push image to public repo (for instance `quay.io`)

```sh
make push IMAGE=quay.io/myorg/apicast-operator VERSION=test
```

* Register the APIcast operator CRDs in the OpenShift API Server

```sh
// As a cluster admin
for i in `ls deploy/crds/*_crd.yaml`; do kubectl create -f $i ; done
```

* Create a new Kubernetes namespace (optional)

```sh
export NAMESPACE=operator-test
kubectl create namespace ${NAMESPACE}

Do not forget to change the current Kubernetes context to the newly
created namespace or create a new context for it and set the new one as the
active context
```

* Deploy the needed roles and ServiceAccounts

```sh
kubectl create -f deploy/service_account.yaml
kubectl create -f deploy/role.yaml
kubectl create -f deploy/role_binding.yaml
```

* Deploy the APIcast operator

```sh
sed -i 's|REPLACE_IMAGE|quay.io/myorg/apicast-operator:test|g' deploy/operator.yaml
kubectl create -f deploy/operator.yaml
```

#### Cleanup manually deployed operator

* Delete all `apicast` custom resources:

```sh
kubectl delete apicasts --all
```

* Delete the APIcast operator, its associated roles and service accounts

```sh
kubectl delete -f deploy/operator.yaml
kubectl delete -f deploy/role_binding.yaml
kubectl delete -f deploy/service_account.yaml
kubectl delete -f deploy/role.yaml
```

* Delete the APIcast CRD:

```sh
kubectl delete crds apicasts.apps.3scale.net
```

### Deploy custom APIcast Operator using OLM

To install this operator on OpenShift 4 using OLM for end-to-end testing, 

* [Push an operator bundle into external app registry](#push-an-operator-bundle-into-external-app-registry).

* Create the [Operator Source](https://github.com/operator-framework/community-operators/blob/master/docs/testing-operators.md#4-create-the-operatorsource)
provided in `deploy/olm-catalog/apicast-operatorsource.yaml` to load your operator bundle in OpenShift.

```bash
kubectl create -f deploy/olm-catalog/apicast-operatorsource.yaml
```

It will take a few minutes for the operator to become visible under
the _OperatorHub_ section of the OpenShift console _Catalog_. It can be
easily found by filtering the provider type to _Custom_.

[git_tool]:https://git-scm.com/downloads
[operator-sdk]:https://github.com/operator-framework/operator-sdk
[go]:https://golang.org/
[kubernetes]:https://kubernetes.io/
[kubectl]:https://kubernetes.io/docs/tasks/tools/install-kubectl/

### Run tests

#### Run unittests

No access to a Kubernetes cluster is required

```sh
make test-crds
```

#### Run integration tests

Access to a Kubernetes v1.11.0+ cluster required

* Run tests locally deploying image
```sh
export NAMESPACE=operator-test
make e2e-run
```

* Run tests locally running operator with go run instead of as an image in the cluster
```sh
export NAMESPACE=operator-test
make e2e-local-run
```

## Manifest management

`operator-courier` is used for metadata syntax checking and validation.
This can be installed directly from pip:

```sh
pip3 install operator-courier
```

### Verify operator manifest

Check [Required fields within your CSV](https://github.com/operator-framework/community-operators/blob/master/docs/required-fields.md)

`operator-courier` will verify the fields included in the Operator metadata (CSV)

```sh
make verify-manifest
```

### Push an operator bundle into external app registry

* Get quay token

Detailed information on this [guide](https://github.com/operator-framework/operator-courier/#authentication)

```bash
curl -sH "Content-Type: application/json" -XPOST https://quay.io/cnr/api/v1/users/login -d '{"user": {"username": "YOURUSERNAME", "password": "YOURPASSWORD"}}' | jq '.token'
```

* Push bundle to Quay.io

Detailed information on this [guide](https://github.com/operator-framework/community-operators/blob/master/docs/testing-operators.md#push-to-quayio).

```bash
make push-manifest APPLICATION_REPOSITORY_NAMESPACE=YOUR_QUAY_NAMESPACE MANIFEST_RELEASE=1.0.0 TOKEN=YOUR_TOKEN
```

## Licenses management

It is a requirement that a file describing all the licenses used in the product is included,
so that users can examine it.

* Check licenses when dependencies change.

```sh
make licenses-check
```

* Update `licenses.xml` file.

```sh
make licenses.xml
```

### Manually adding a new license

When licenses check does not parse correctly licensing information, it will complain.
In that case, you need to add manually license information.

There are two options: a)specify dependency license (recommended) or b)add exception for that dependency.

* Specify dependency license:

```sh
license_finder dependencies add YOURLIBRARY --decisions-file=doc/dependency_decisions.yml LICENSE --project-path "PROJECT URL"
```

For instance

```sh
license_finder dependencies add k8s.io/klog --decisions-file=doc/dependency_decisions.yml "Apache 2.0" --project-path "https://github.com/kubernetes/klog"
```

* Adding exception for a dependency:

```sh
license_finder approval add YOURLIBRARY --decisions-file=doc/dependency_decisions.yml --why "LICENSE_TYPE LINK_TO_LICENSE"
```

For instance

```sh
license_finder approval add github.com/golang/glog --decisions-file=doc/dependency_decisions.yml --why "Apache 2.0 License https://github.com/golang/glog/blob/master/LICENSE"
```
