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
  * [(re)Generate an operator bundle image](#(re)generate-an-operator-bundle-image)
  * [Validate an operator bundle image](#validate-an-operator-bundle-image)
  * [Push an operator bundle into an external container repository](#push-an-operator-bundle-into-an-external-container-repository)
* [Licenses management](#licenses-management)
  * [Manually adding a new license](#manually-adding-a-new-license)

## Prerequisites

* [operator-sdk] version v1.1.0
* [docker] version 17.03+
* [git][git_tool]
* [go] version 1.13+
* [kubernetes] version v1.11.3+
* [kubectl] version v1.11.3+
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
make docker-build IMG=quay.io/myorg/apicast-operator:myversiontag
```

## Run APIcast Operator

### Run APIcast Operator Locally

Run the operator locally with the default Kubernetes config file present at $HOME/.kube/config

Run operator from command line, it will not be deployed as pod.

* [Clone the repo](#Clone-repository)

* Register the APIcast operator CRD in the Kubernetes API Server

```sh
// As a cluster admin
for i in `ls bundle/manifests/**apps.3scale.net_*.yaml`; do kubectl create -f $i ; done
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

To install this operator on an OpenShift 4.5+ cluster using OLM for
end-to-end testing:

* Perform naming changes to avoid collision with existing APIcast Operator
  official public operators catalog entries:
  * Edit the `bundle/manifests/apicast-operator.clusterserviceversion.yaml` file
    and perform the following changes:
      * Change the current value of `.metadata.name` to a different name
        than `apicast-operator.v*`. For example to `myorg-apicast-operator.v0.0.1`
      * Change the current value of `.spec.displayName` to a value that helps you
        identify the catalog entry name from other operators and the official
        APIcast operator entries. For example to `"MyOrg apicast operator"`
      * Change the current value of `.spec.provider.Name` to a value that helps
        you identify the catalog entry name from other operators and the official
        APIcast operator entries. For example, to `MyOrg`
  * Edit the `bundle.Dockerfile` file and change the value of
    the Dockerfile label `LABEL operators.operatorframework.io.bundle.package.v1`
    to a different value than `apicast-operator`. For example to
    `myorg-apicast-operator`
  * Edit the `bundle/metadata/annotations.yaml` file and change the value of
    `.annotations.operators.operatorframework.io.bundle.package.v1` to a
    different value than `apicast-operator`. For example to
    `myorg-apicast-operator`. The new value should match the
    Dockerfile label `LABEL operators.operatorframework.io.bundle.package.v1`
    in the `bundle.Dockerfile` as explained in the point above

  It is really important that all the previously shown fields are changed
  to avoid overwriting the APIcast operator official public operator
  catalog entry in your cluster and to avoid confusion having two equal entries
  on it.

* [Create an operator bundle image](#(re)Generate-operator-bundle) using the
  changed contents above

* [Push the operator bundle into an external container repository](#push-an-operator-bundle-into-an-external-container-repository).

* Run the following command to deploy the operator in your currently configured
  and active cluster in $HOME/.kube/config:
  ```sh
  operator-sdk run bundle --namespace <mynamespace>
  ```

  Additionally, a specific kubeconfig file with a desired Kubernetes
  configuration can be provided too:
  ```sh
  operator-sdk run bundle --namespace <mynamespace> --kubeconfig <path>
  ```

It will take a few minutes for the operator to become visible under
the _OperatorHub_ section of the OpenShift console _Catalog_. It can be
easily found by filtering the provider type to _Custom_.

[git_tool]:https://git-scm.com/downloads
[operator-sdk]:https://github.com/operator-framework/operator-sdk
[docker]:https://docs.docker.com/install/
[go]:https://golang.org/
[kubernetes]:https://kubernetes.io/
[kubectl]:https://kubernetes.io/docs/tasks/tools/install-kubectl/

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
make test-e2e
```

## Bundle management

### (re)Generate an operator bundle image

```sh
make bundle
```

The generated output will be saved in the `bundle` directory

### Validate an operator bundle image

```sh
make bundle-validate
```


### Push an operator bundle into an external container repository

```sh
make docker-push IMG=quay.io/myorg/apicast-operator:myversiontag
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
