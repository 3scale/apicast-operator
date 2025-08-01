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
    * [Setting Horizontal Pod Autoscaling](#setting-horizontal-pod-autoscaling)
    * [Enabling TLS at pod level](#enabling-tls-at-pod-level)
    * [Adding custom policies](adding-custom-policies.md)
    * [Adding custom environments](adding-custom-environments.md)
    * [Gateway instrumentation](gateway-instrumentation.md)
* [Reconciliation](#reconciliation)
* [Upgrading APIcast](#upgrading-APIcast)
* [APIcast CRD reference](apicast-crd-reference.md)
  * CR samples [\[1\]](../config/samples/apps_v1alpha1_apicast_admin_portal_url_cr.yaml) [\[2\]](cr_samples/)

## Installing APIcast self-managed gateway

This section will take you through installing and deploying an APIcast self-managed
gateway solution solution via the APIcast operator,
using the [*APIcast*](apicast-crd-reference.md) custom resource.

Deploying the APIcast custom resource will make the operator begin processing
and will deploy an APIcast self-managed gateway solution from it.

### Prerequisites

* OpenShift Container Platform 4.6 or newer, or a Kubernetes 1.19 native installation
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

**Enabling TLS terminator in the ingress object**

TLS for the exposedHost section can also be configured optionally.

Steps are:

1.- Optionally generate self signed certificates for your DOMAIN
```
openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout server.key -out server.crt
```

Fill out the prompts appropriately. The most important line is the one that requests the Common Name (e.g. server FQDN or YOUR name).
You need to enter the domain name associated with your server or, more likely, your server’s public IP address.

2.- Create the certificate secret
```
kubectl create secret tls mycertsecret --cert=server.crt --key=server.key
```

**Watch for secret changes**

By default, content changes in the secret will not be noticed by the apicast operator.
The apicast operator allows monitoring the secret for changes adding the `apicast.apps.3scale.net/watched-by=apicast` label.
With that label in place, when the content of the secret is changed, the operator will get notified.
Then, the operator will rollout apicast deployment to make the changes effective.
The operator will not take *ownership* of the secret in any way.

```
kubectl label secret ${SOME_SECRET_NAME} apicast.apps.3scale.net/watched-by=apicast
```

3.- Reference the certificate secret in APIcast CR

```
apiVersion: apps.3scale.net/v1alpha1
kind: APIcast
metadata:
  name: apicast1
spec:
  ...
  exposedHost:
    host: example.com
    tls:
    - hosts:
      - example.com
      secretName: mycertsecret
```

**[For Openshift users]** Starting with OCP 4.6, you can use the default cluster ingress certificate. The APIcast CR to be used would be:

```
apiVersion: apps.3scale.net/v1alpha1
kind: APIcast
metadata:
  name: apicast1
spec:
  ...
  exposedHost:
    host: example.com
    tls:
    - {}
```

**[For Openshift users]** Starting with OCP 4.12, an alert has been added for Ingress object without IngressClassName. To silence
the alert, you can specify the ingressClassName:

```
apiVersion: apps.3scale.net/v1alpha1
kind: APIcast
metadata:
  name: apicast1
spec:
  ...
  exposedHost:
    host: example.com
    ingressClassName: openshift-default
    tls:
    - {}
```

Details about the available fields in the `exposedHost` section can be found [here](apicast-crd-reference.md#APIcastExposedHost)

#### Setting Horizontal Pod Autoscaling 
Horizontal Pod Autoscaling(HPA) is available for Apicasts. To enable HPA set the apicast.spec.hpa to `true`. HPA will be created with default values.

- replicas: min 1; max 5;
- request resource requirements: cpu: 1000m; memory: 128Mi;
- limits resource requirements: cpu: 1000m; memory: 128Mi;

HPA object can be edited and the operator will not revert changes.

The following is an example of the output HPA using the defaults. 

```yaml
kind: HorizontalPodAutoscaler
apiVersion: autoscaling/v2
metadata:
  name: example-apicast
spec:
  scaleTargetRef:
    kind: Deployment
    name: apicast-example-apicast
    apiVersion: apps/v1
  minReplicas: 1
  maxReplicas: 5
  metrics:
    - type: Resource
      resource:
        name: memory
        target:
          type: Utilization
          averageUtilization: 85
    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: 85
```
Here is an example of the Apicast CR set with the HPA on using the default values:

```yaml
apiVersion: apps.3scale.net/v1alpha1
kind: APIcast
metadata:
  name: example-apicast
  namespace: apicast
spec:
  adminPortalCredentialsRef:
    name: <Admin portal credentials reference>
  hpa: true
```
Removing hpa field or setting enabled to false will remove the HPA for the component and the specified within spec 
replicas will be applied, if not specified, the operator will set its default values. 

You can still scale vertically by setting the resource requirements for Apicast. As HPA scales on 85% of requests
values having extra resources set aside for limits is unnecessary i.e. set your requests equal to your limits when scaling
vertically.

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

#### Enabling TLS at pod level

You can use your SSL certificate to enable TLS at APIcast pod level setting either `httpsPort` or `httpsCertificateSecretRef` fields or both.


Steps to enable TLS at pod level:

1.- Optionally generate self signed certificates for your DOMAIN
```
openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout server.key -out server.crt
```

Fill out the prompts appropriately. The most important line is the one that requests the Common Name (e.g. server FQDN or YOUR name).
You need to enter the domain name associated with your server or, more likely, your server’s public IP address.

2.- Create the certificate secret
```
kubectl create secret tls mycertsecret --cert=server.crt --key=server.key
```

3.- Reference the certificate secret in APIcast CR

```
apiVersion: apps.3scale.net/v1alpha1
kind: APIcast
metadata:
  name: apicast1
spec:
  ...
  httpsPort: 8443
  httpsCertificateSecretRef:
    name: mycertsecret
```

**NOTE 1**: If `httpsPort` is set and `httpsCertificateSecretRef` is not set, APIcast will use a default certificate bundled in the image.

**NOTE 2**: If `httpsCertificateSecretRef` is set and `httpsPort` is not set, APIcast will enable TLS at port number `8443`.

See [APIcast CRD reference](apicast-crd-reference.md)

The TLS port can be accessed using apicast service's named port `httpsproxy`. You can check using `kubectl port-forward` command.

Open a terminal and run the port forwarding command for `httpsproxy` named port.
```
$ kubectl port-forward service/apicast-apicast1 httpsproxy
Forwarding from 127.0.0.1:8443 -> 8443
Forwarding from [::1]:8443 -> 8443
```

In other terminal, download used certificate.
```
$ echo quit | openssl s_client -showcerts -connect 127.0.0.1:8443 2>/dev/null | sed -ne '/-BEGIN CERTIFICATE-/,/-END CERTIFICATE-/p'
```

The downloaded certificate should match provided certificate.

#### Override default CA certificate at pod level
You can override the default CA certificate used by APIcast pod with `caCertificateSecretRef` field.

Steps to override CA certificate at pod level:

1.- Genrate CA certificate
```
openssl genrsa -out rootCA.key 2048
openssl req -batch -new -x509 -nodes -key rootCA.key -sha256 -days 1024 -out rootCA.pem
```

2.- Create the certificate secret
```
kubectl create secret generic cacert --namespace=apicast-test --from-file=ca-bundle.crt=rootCA.pem
```

3.- Reference the certificate secret in APIcast CR

```
apiVersion: apps.3scale.net/v1alpha1
kind: APIcast
metadata:
  name: apicast1
spec:
  ...
  caCertificateSecretRef:
    name: cacert
```

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
