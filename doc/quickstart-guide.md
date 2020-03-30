# Quickstart Guide

## Table of contents

* [Install the APIcast operator](#Install-the-APIcast-operator)
  * [In an OpenShift environment using the OpenShift catalog](#In-OpenShift-using-the-OpenShift-catalog)
  * [In a Kubernetes native environment using OperatorHub.io](#In-a-Kubernetes-native-environment-using-OperatorHub\.io)
* [Deploying an APIcast gateway self-managed solution using the operator](#Deploying-an-APIcast-gateway-self-managed-solution-using-the-operator)
  * [Providing a 3scale Porta endpoint](#Providing-a-3scale-Porta-endpoint)
  * [Providing a configuration Secret](#Providing-a-configuration-Secret)

## Install the APIcast operator

### In OpenShift using the OpenShift catalog

You will need access to an OpenShift Container Platform 4.1 cluster or newer
to use this installation method.

Procedure
1. In the OpenShift Container Platform console, log in using an account with administrator privileges
1. Create new project `operator-test` in *Projects > Create Project*
1. Click *Catalog > OperatorHub*
1. In the Filter by keyword box, type `apicast` to find the APIcast operator
1. Click the APIcast operator. Information about the Operator is displayed
1. Click *Install*. The Create Operator Subscription page opens
1. On the *Create Operator Subscription* page, accept all of the default selections and click Subscribe
1. After the subscription *upgrade status* is shown as *Up to date*,
   click *Catalog > Installed Operators* to verify that the APIcast operator
   ClusterServiceVersion (CSV) is displayed and its Status ultimately resolves
   to _InstallSucceeded_ in the `operator-test` project

### In a Kubernetes native environment using OperatorHub.io

You will need access to a Kubernetes 1.11 cluster or newer to use this
installation method.

Procedure
1. Access to the [OperatorHub.io](https://operatorhub.io/) website
1. In the filter search bar type `apicast` to find the APIcast operator
1. Click the APIcast operator. Information about the Operaotr is displayed
1. Click the *Install* button. A new window with the instructions to install
   the operator will appear. Follow them to install the operator. This will
   install the operator in a namespace called `my-apicast-community-operator`
1. After installing it you can verify that the APIcast operator comes up by
   verifying its ClusterServiceVersion (CSV) is displayed and its Status
   ultimately resolves to _InstallSucceeded_ in the
   `my-apicast-community-operator` namespace by executing
   ```
   kubectl get csv -n my-apicast-community-operator -o yaml
   ```

## Deploying an APIcast gateway self-managed solution using the operator

Deploying the *APIcast* custom resource will make the operator begin processing
and will deploy an APIcast gateway self-managed solution solution from it.

To deploy an APIcast custom resource the kubectl command line tool can be used.
Alternatively, if the APIcast operator has been installed in an OpenShift
installation using Operator Lifecycle Manager (OLM) you can install an APIcast
gateway self-managed solution using the UI by:
  1. Using the OpenShift Container Platform console, Click *Catalog > Installed Operators*. From the list
   of *Installed Operators*, click APIcast Operator.
  1. Click *APIcast > Create APIcast*

APIcast gateway self-managed can be deployed and configured using two main approaches:
   * [Providing a 3scale Porta endpoint](#Providing-a-3scale-Porta-endpoint)
   * [Providing a configuration Secret](#Providing-a-configuration-Secret)

### Providing a 3scale Porta endpoint

Set the following content when creating the APIcast object:

```
apiVersion: apps.3scale.net/v1alpha1
kind: APIcast
metadata:
  name: example-apicast
spec:
  adminPortalCredentialsRef:
    name: asecretname
```

The ```spec.adminPortalCredentialsRef.name``` must be the name of an
already existing Kubernetes secret that contains a 3scale [Porta](https://github.com/3scale/porta/)
admin portal endpoint information. For more information about the contents of the secret
see the [Admin portal configuration secret](apicast-crd-reference.md#AdminPortalSecret) reference.

After creating the APIcast object you should verify that the APIcast pod is
running and ready.

To do so verify that the `readyReplicas` field of the Kubernetes Deployment
associated to the APIcast object is 1 or wait until it is:

```
$ echo $(kubectl get deployment apicast-example-apicast -o jsonpath='{.status.readyReplicas}')
1
```

#### Verify APIcast gateway is running and available

To verify that APIcast gateway is running and available you can port-forward
the Kubernetes Service exposed by APIcast to your local machine and perform
a test request.

To do so:

Port-forward the APIcast Kubernetes Service to localhost:8080
```
kubectl port-forward svc/apicast-example-apicast 8080
```

Then you can make a request to one of the 3scale Services that have been
configured in 3scale and verify that a successful HTTP response with the expected
contents of the configured backend for the corresponding specific service that
is being checked.

For example, if the service that has been created has been configured with
the host "myhost.com" then the following can be executed:

```
$  curl 127.0.0.1:8080/test -H "Host: myhost.com"
```

A successful HTTP response should be received with the expected content
that the Service backend should provide.

### Providing a configuration Secret

Set the following content when creating the APIcast object:

```
apiVersion: apps.3scale.net/v1alpha1
kind: APIcast
metadata:
  name: example-apicast
spec:
  embeddedConfigurationSecretRef:
    name: asecretname
```

The ```spec.embeddedConfigurationSecretRef.name``` must be the name of an
already existing Kubernetes secret that contains the configuration of the
gateway. For more information about the contents of the secret and the
configuration of the gateway see the
[Embedded configuration secret](apicast-crd-reference.md#EmbeddedConfSecret) reference.

An example of an embedded configuration secret that configures a 3scale service
with the hostname "localhost" with the
[3scale echo API](https://github.com/3scale/echo-api/) as the corresponding
backend of the 3scale service:

```
apiVersion: v1
kind: Secret
metadata:
  name: asecretname
type: Opaque
stringData:
  config.json: |
    {
      "services": [
        {
          "proxy": {
            "policy_chain": [
              { "name": "apicast.policy.upstream",
                "configuration": {
                  "rules": [{
                    "regex": "/",
                    "url": "http://echo-api.3scale.net"
                  }]
                }
              }
            ]
          }
        }
      ]
    }
```

After creating the APIcast object you should verify that the APIcast pod is
running and ready.

To do so verify that the `readyReplicas` field of the Kubernetes Deployment
associated to the APIcast object is 1 or wait until it is:

```
$ echo $(kubectl get deployment apicast-example-apicast -o jsonpath='{.status.readyReplicas}')
1
```

#### Verify APIcast gateway is running and available

To verify that APIcast gateway is running and available you can port-forward
the Kubernetes Service exposed by APIcast to your local machine and perform
a test request.

To do so:

Port-forward the APIcast Kubernetes Service to localhost:8080
```
kubectl port-forward svc/apicast-example-apicast 8080
```

Then you can make a request to the 3scale Service and verify that you get a
successful HTTP response with the contents of the configured backend for
the service. In this case, the contents of the echo API URL.

For example:

```
$  curl 127.0.0.1:8080/test -H "Host: localhost"
{
  "method": "GET",
  "path": "/test",
  "args": "",
  "body": "",
  "headers": {
    "HTTP_VERSION": "HTTP/1.1",
    "HTTP_HOST": "echo-api.3scale.net",
    "HTTP_ACCEPT": "*/*",
    "HTTP_USER_AGENT": "curl/7.65.3",
    "HTTP_X_REAL_IP": "127.0.0.1",
    "HTTP_X_FORWARDED_FOR": ...
    "HTTP_X_FORWARDED_HOST": "echo-api.3scale.net",
    "HTTP_X_FORWARDED_PORT": "80",
    "HTTP_X_FORWARDED_PROTO": "http",
    "HTTP_FORWARDED": "for=10.0.101.216;host=echo-api.3scale.net;proto=http"
  },
  "uuid": "603ba118-8f2e-4991-98c0-a9edd061f0f0"
```
