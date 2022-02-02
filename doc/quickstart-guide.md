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
1. Create a new project `operator-test` in *Projects > Create Project*
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

You will need access to a Kubernetes 1.19 cluster or newer to use this
installation method.

Procedure
1. Access to the [OperatorHub.io](https://operatorhub.io/) website
1. In the filter search bar type `apicast` to find the APIcast operator
1. Click the APIcast operator. Information about the Operator is displayed
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
and will deploy an APIcast gateway self-managed solution from it.

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

1. Create a kubernetes secret that contains a 3scale Porta admin portal endpoint information

```
kubectl create secret generic ${SOME_SECRET_NAME} --from-literal=AdminPortalURL=MY_3SCALE_URL
```

`${SOME_SECRET_NAME}` is the name of the secret and can be any name you want, provided it does not conflict with any other existing secret.

`${MY_3SCALE_URL}` is the URI that includes your password and 3scale Porta portal endpoint. See [format](https://github.com/3scale/APIcast/blob/master/doc/parameters.md#threescale_portal_endpoint)

Example:

```
kubectl create secret generic 3scaleportal --from-literal=AdminPortalURL=https://access-token@account-admin.3scale.net
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

For more information about the contents of the secret see the [Admin portal configuration secret reference](apicast-crd-reference.md#AdminPortalSecret).

2. Create APIcast object:

```
apiVersion: apps.3scale.net/v1alpha1
kind: APIcast
metadata:
  name: example-apicast
spec:
  adminPortalCredentialsRef:
    name: asecretname
```

The `spec.adminPortalCredentialsRef.name` must be the name of the previously created secret.

After creating the APIcast object you should verify that the APIcast pod is
running and ready.

To do so verify that the `readyReplicas` field of the Kubernetes Deployment
associated with the APIcast object is 1 or wait until it is:

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

Make a request to a configured 3scale product API to verify a successful HTTP response.
Use the domain name configured in `Staging Public Base URL` or `Production Public Base URL` settings of your product.

For example, if the product that has been created has been configured with
the host "myhost.com" then the following can be executed:

```
curl 127.0.0.1:8080/test -H "Host: myhost.com"
```

A successful HTTP response should be received with the expected content
that the Service backend should provide.

### Providing a configuration Secret

1. Create a kubernetes secret that contains the gateway embedded configuration

An example of an embedded configuration secret that configures a 3scale service
with the hostname "localhost" with the
[3scale echo API](https://github.com/3scale/echo-api/) as the corresponding
backend of the 3scale product:

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

For more information about the contents of the secret and the
configuration of the gateway sees the
[Embedded configuration secret](apicast-crd-reference.md#EmbeddedConfSecret) reference.

**Watch for secret changes**

By default, content changes in the secret will not be noticed by the apicast operator.
The apicast operator allows monitoring the secret for changes adding the `apicast.apps.3scale.net/watched-by=apicast` label.
With that label in place, when the content of the secret is changed, the operator will get notified.
Then, the operator will rollout apicast deployment to make the changes effective.
The operator will not take *ownership* of the secret in any way.

```
kubectl label secret ${SOME_SECRET_NAME} apicast.apps.3scale.net/watched-by=apicast
```

2. Create APIcast object:

```
apiVersion: apps.3scale.net/v1alpha1
kind: APIcast
metadata:
  name: example-apicast
spec:
  embeddedConfigurationSecretRef:
    name: asecretname
```

The `spec.embeddedConfigurationSecretRef.name` must be the name of the previously created secret.

After creating the APIcast object you should verify that the APIcast pod is
running and ready.

To do so verify that the `readyReplicas` field of the Kubernetes Deployment
associated with the APIcast object is 1 or wait until it is:

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

Then you can make a request and verify that you get a
successful HTTP response from the echo API.

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
