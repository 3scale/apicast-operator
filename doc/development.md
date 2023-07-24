# Development Guide

## Table of contents
* [Prerequisites](#Prerequisites)
* [Clone repository](#Clone-repository)
* [Building APIcast operator image](#building-APIcast-operator-image)
* [Run APIcast Operator](#run-APIcast-operator)
  * [Run APIcast Operator Locally](#run-APIcast-operator-locally)
  * [Deploy custom APIcast Operator using OLM](#deploy-custom-APIcast-operator-using-olm)
* [Run tests](#run-tests)
  * [Run all tests](#run-all-tests)
  * [Run unit tests](#run-unit-tests)
  * [Run end-to-end tests](#run-end-to-end-tests)
* [Bundle management](#bundle-management)
  * [Generate an operator bundle image](#generate-an-operator-bundle-image)
  * [Validate an operator bundle image](#validate-an-operator-bundle-image)
  * [Push an operator bundle into an external container repository](#push-an-operator-bundle-into-an-external-container-repository)
* [Licenses management](#licenses-management)
  * [Manually adding a new license](#manually-adding-a-new-license)

## Prerequisites

* [operator-sdk] version v1.2.0
* [docker] version 17.03+
* [git][git_tool]
* [go] version 1.19
* [kubernetes] version v1.22.1+
* [kubectl] version v1.22.0+
* Access to a Kubernetes v1.21.0+ cluster.
* A user with administrative privileges in the Kubernetes cluster.
* Make sure that the `DOCKER_ORG` and `DOCKER_REGISTRY` environment variables are set to the same value as
  your username on the container registry, and the container registry you are using.

```sh
export DOCKER_ORG=docker_hub_username
export DOCKER_REGISTRY=quay.io
```

## Clone repository

```sh
git clone https://github.com/3scale/apicast-operator
cd apicast-operator
```
## Building APIcast operator image

Build operator image

```sh
make docker-build-only IMG=$DOCKER_REGISTRY/$DOCKER_ORG/apicast-operator:myversiontag
```

## Run APIcast Operator

### Run APIcast Operator Locally

Run the operator locally with the default Kubernetes config file present at $HOME/.kube/config

Run operator from command line, it will not be deployed as pod.

* Register the APIcast operator CRD in the Kubernetes API Server

```sh
// As a cluster admin
make install
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
make download
```

* Run operator

```sh
make run
```

### Deploy custom APIcast Operator using OLM

* Build and upload custom operator image
```
make docker-build-only IMG=$DOCKER_REGISTRY/$DOCKER_ORG/apicast-operator:myversiontag
make operator-image-push IMG=$DOCKER_REGISTRY/$DOCKER_ORG/apicast-operator:myversiontag
```

* Build and upload custom operator bundle image. Changes to avoid conflicts will be made by the makefile.
```
make bundle-custom-build IMG=$DOCKER_REGISTRY/$DOCKER_ORG/apicast-operator:myversiontag BUNDLE_IMG=$DOCKER_REGISTRY/$DOCKER_ORG/apicast-operator-bundles:myversiontag
make bundle-image-push BUNDLE_IMG=$DOCKER_REGISTRY/$DOCKER_ORG/apicast-operator-bundles:myversiontag
```

* Deploy the operator in your currently configured and active cluster in $HOME/.kube/config:
```
make bundle-run BUNDLE_IMG=$DOCKER_REGISTRY/$DOCKER_ORG/apicast-operator-bundles:myversiontag
```

**Note**: The _catalogsource_ will be installed in the `openshift-marketplace` namespace
[issue](https://bugzilla.redhat.com/show_bug.cgi?id=1779080). By default, cluster scoped
subscription will be created in the namespace `openshift-marketplace`.
Feel free to delete the operator (from the UI **OperatorHub -> Installed Operators**)
and install it namespace or cluster scoped.

It will take a few minutes for the operator to become visible under
the _OperatorHub_ section of the OpenShift console _Catalog_. It can be
easily found by filtering the provider type to _Custom_.

### Run tests

#### Run all tests

```sh
make test
```

#### Run unit tests

```sh
make test-unit
```

#### Run end-to-end tests

```sh
make kind-create-cluster
make test-integration
```

## Bundle management

### Generate an operator bundle image

```sh
make bundle-build BUNDLE_IMG=$DOCKER_REGISTRY/$DOCKER_ORG/myrepo:myversiontag
```

### Push an operator bundle into an external container repository

```sh
make bundle-image-push BUNDLE_IMG=$DOCKER_REGISTRY/$DOCKER_ORG/myrepo:myversiontag
```

### Validate an operator bundle image

NOTE: if validating an image, the image must exist in a remote registry, not just locally.

```sh
make bundle-validate-image BUNDLE_IMG=$DOCKER_REGISTRY/$DOCKER_ORG/myrepo:myversiontag
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

[git_tool]:https://git-scm.com/downloads
[operator-sdk]:https://github.com/operator-framework/operator-sdk
[docker]:https://docs.docker.com/install/
[go]:https://golang.org/
[kubernetes]:https://kubernetes.io/
[kubectl]:https://kubernetes.io/docs/tasks/tools/install-kubectl/
