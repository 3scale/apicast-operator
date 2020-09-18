# User Guide

## Table of contents
* [Installing APIcast self-managed gateway](#installing-apicast-self-managed-gateway)
  * [Prerequisites](#Prerequisites)
  * [Basic Installation](#Basic-installation)
  * [Deployment Configuration Options](#Deployment-Configuration-Options)
    * [Providing the APIcast configuration through an available 3scale Porta endpoint](#Providing-the-APIcast-configuration-through-an-available-3scale-Porta-endpoint)
    * [Providing the APIcast configuration through a configuration file](#Providing-the-APIcast-configuration-through-a-configuration-file)
    * [Exposing APIcast externally via a Kubernetes Ingress](#Exposing-APIcast-externally-via-a-Kubernetes-Ingress)
    * [Setting custom resource requirements](#setting-custom-resource-requirements)
* [Reconciliation](#reconciliation)
* [Upgrading APIcast](#upgrading-APIcast)
* [APIcast CRD reference](apicast-crd-reference.md)

## Installing APIcast self-managed gateway

This section will take you through installing and deploying an APIcast self-managed
gateway solution solution via the APIcast operator,
using the [*APIcast*](apicast-crd-reference.md) custom resource.

Deploying the APIcast custom resource will make the operator begin processing
and will deploy an APIcast self-managed gateway solution from it.

### Prerequisites

* OpenShift Container Platform 4.1 or newer, or a Kubernetes 1.11 native installation
  or newer
* Deploying an APIcast self-managed solution using the operator first requires
  that you follow the steps in [quickstart guide](quickstart-guide.md) about
  [Install the APIcast operator](quickstart-guide.md#Install-the-APIcast-operator)
* In case APIcast gateway self-managed is configured using 3scale Porta an existing
  3scale API management installation is needed and the 3scale Porta admin
  portal endpoint has to be accessible from from where APIcast gateway
  self-managed is installed
* In case APIcast gateway self-managed is configured to be exposed externally
  a default Kubernetes Ingress controller or an existing configured has to be
  available. Look at the [Exposing APIcast externally via a Kubernetes Ingress](#Exposing-APIcast-externally-via-a-Kubernetes-Ingress)
  subsection for details about that

### Basic installation

Follow the [Deploying an APIcast gateway self-managed solution using the operator](quickstart-guide.md#Deploying-an-APIcast-gateway-self-managed-solution-using-the-operator) section in the
[quickstart guide](quickstart-guide.md)

### Deployment Configuration Options

By default, the following deployment configuration options will be applied:
* A Kubernetes Deployment with one replica set associated to the APIcast object
* APIcast will not be exposed externally. No Kubernetes Ingress objects will be created
* A Kubernetes Service configured with a ClusterIP, pointing to the Kubernetes
  Deployment associated to the APIcast object. This service has the
  following ports configured and accessible:
    * 8080/TCP: The APIcast gateway proxy port
    * 8090/TCP: The APIcast gateway management port
* Resource requirements: *CPU* [Request: 500m, Limit: 1], *Memory* [Request: 64Mi, Limit: 128Mi]

Default configuration option is suitable for PoC or evaluation.
One, many or all of the default configuration options can be overriden with
specific field values in the [*APIcast*](apicast-crd-reference.md) custom resource.
#### Providing the APIcast configuration through an available 3scale Porta endpoint

Follow [this](quickstart-guide.md#Providing-a-3scale-Porta-endpoint) section in the [quickstart guide](quickstart-guide.md)

#### Providing the APIcast configuration through a configuration file

Follow [this](quickstart-guide.md#Providing-a-configuration-Secret) section in the [quickstart guide](quickstart-guide.md)

#### Exposing APIcast externally via a Kubernetes Ingress

To do so, the `exposedHost` section can be set and configured.

When the `host` field in the `exposedHost` section is set a Kubernetes Ingress
object is created. This Kubernetes Ingress object then can be used by a
previously installed and existing Kubernetes Ingress Controller to make APIcast
accessible externally.

To learn what Ingress Controllers are available to make APIcast externally accessible
and how they are configured see the
[Kubernetes Ingress Controllers documentation](https://kubernetes.io/docs/concepts/services-networking/ingress-controllers/)

For example, to expose APIcast with the hostname `myhostname.com`:

```yaml
apiVersion: apps.3scale.net/v1alpha1
kind: APIcast
metadata:
  name: example-apicast
spec:
  ...
  exposedHost:
    host: "myhostname.com"
  ...
```

This will create a Kubernetes Ingress object on the port 80 using HTTP.
In case APIcast has been deployed in an OpenShift environment, by default
the OpenShift default Ingress Controller will create a Route object using
the Ingress object created by APIcast to allow external access to the APIcast
installation.

TLs for the exposedHost section can also be configured if desired. Details
about the available fields in the `exposedHost` section can be
found [here](apicast-crd-reference.md#APIcastExposedHost)

#### Setting custom resource requirements

Default [Resource Requirements](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/)
can be customized to suit your needs via Apicast Custom Resource `resources` attribute field.

For example, setting custom resource requirements:

```yaml
apiVersion: apps.3scale.net/v1alpha1
kind: APIcast
metadata:
  name: apicast1
spec:
  resources:
    requests:
      memory: "150Mi"
      cpu: "300m"
    limits:
      memory: "500Mi"
      cpu: "2000m"
```

Two notes:
* When resource requirements are not specified, the operator [defaults](#deployment-configuration-options) are being set.
* When resource requests and/or resource limits are not specified, the operator [defaults](#deployment-configuration-options) will *NOT* be used, instead no requests and/or limit will be set.

See [APIcast CRD reference](apicast-crd-reference.md) 

### Reconciliation
After an APIcast self-managed gateway solution has been installed, APIcast
operator enables updating a given set of parameters from the custom resource
in order to modify APIcast configuration options. Modifications are performed
in a hot swapping way, i.e., without stopping or shutting down the system.

### Upgrading APIcast
Upgrading an APIcast self-managed gateway solution requires upgrading
the APIcast operator. However, upgrading the APIcast operator does not
necessarily imply upgrading an APIcast self-managed gateway solution. The
operator could be upgraded because there are bugfixes or security fixes.

The recommended way to upgrade the APIcast operator is via the Operator Lifecycle Manager (OLM).

If you selected *Automatic updates* when APIcast operator was installed,
when a new version of the operator is available, the Operator Lifecycle
Manager (OLM) automatically upgrades the running instance of the operator
without human intervention.

If you selected *Manual updates*, when a newer version of the Operator is
available, the OLM creates an update request. As a cluster administrator, you
must then manually approve that update request to have the Operator updated
to the new version.
