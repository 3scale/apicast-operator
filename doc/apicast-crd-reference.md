# APIcast Custom Resource reference

**json/yaml field**| **Type** | **Required** | **Default value** | **Description** |
| --- | --- | --- | --- | --- |
| `spec` | [APIcastSpec](#APIcastSpec) | Yes | See [APIcastSpec](#APIcastSpec) | The specfication for Apicast custom resource |
| `status` | [APIcastStatus](#APIcastStatus) | No | See [APIcastStatus](#APIcastStatus) | The status for the custom resource  |

#### APIcastSpec

**json/yaml field**| **Type** | **Required** | **Default value** | **Description** |
| --- | --- | --- | --- | --- |
| `replicas` | integer | No | 1 | Number of replica pods |
| `adminPortalCredentialsRef` | LocalObjectReference | No | N/A | Secret with the portal endpoint URL information. See [AdminPortalSecret](#AdminPortalSecret) for required format |
| `embeddedConfigurationSecretRef` | LocalObjectReference | No | N/A | Secret containing the gateway configuration. See [EmbeddedConfSecret](#EmbeddedConfSecret) for required format |
| `serviceAccount` | string | No | `default` service account | Service account associated to the gateway |
| `image` | string | No | Official apicast image | Apicast gateway container image. Only for devtesting purposes |
| `exposedHost` | [APIcastExposedHost](#APIcastExposedHost) | No | No external access | Domain name used for external access |
| `deploymentEnvironment` | string | No | N/A | Environment for which the configuration (see [docs](https://github.com/3scale/APIcast/blob/master/doc/parameters.md#threescale_deployment_env)) |
| `dnsResolverAddress` | string | No | N/A | DNS resolver (see [docs](https://github.com/3scale/APIcast/blob/master/doc/parameters.md#resolver)) |
| `enabledServices` | []string | No | N/A | List of service IDs used to filter the services configured (see [docs](https://github.com/3scale/APIcast/blob/master/doc/parameters.md#apicast_services_list)) |
| `configurationLoadMode` | string | No | N/A | Defines how to load the configuration (see [docs](https://github.com/3scale/APIcast/blob/master/doc/parameters.md#apicast_configuration_loader)) |
| `logLevel` | string | No | N/A | Log level for the OpenResty logs  (see [docs](https://github.com/3scale/APIcast/blob/master/doc/parameters.md#apicast_log_level)) |
| `pathRoutingEnabled` | bool | No | N/A | When this parameter is set to true, the gateway will use path-based routing in addition to the default host-based routing (see [docs](https://github.com/3scale/APIcast/blob/master/doc/parameters.md#apicast_path_routing)) |
| `responseCodesIncluded` | bool | No | N/A | When set to true, APIcast will log the response code of the response returned by the API backend in 3scale (see [docs](https://github.com/3scale/APIcast/blob/master/doc/parameters.md#apicast_response_codes)) |
| `cacheConfigurationSeconds` | integer | No | N/A | Specifies the period (in seconds) that the configuration will be stored in the cache (see [docs](https://github.com/3scale/APIcast/blob/master/doc/parameters.md#apicast_configuration_cache)) |
| `managementAPIScope` | string | No | N/A | Apicast management API configuration control (see [docs](https://github.com/3scale/APIcast/blob/master/doc/parameters.md#apicast_management_api)) |
| `openSSLPeerVerificationEnabled` | bool | No | N/A | Controls the OpenSSL Peer Verification (see [docs](https://github.com/3scale/APIcast/blob/master/doc/parameters.md#openssl_verify)) |
| `resources` | [v1.ResourceRequirements](https://v1-17.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#resourcerequirements-v1-core) | No | *CPU* [Request: 500m, Limit: 1], *Memory* [Request: 64Mi, Limit: 128Mi] | Resources describes the compute resource requirements |
| `upstreamRetryCases` | string | No | N/A | Specifies in which cases a request to the upstream API should be retried (see [docs](https://github.com/3scale/APIcast/blob/master/doc/parameters.md#apicast_upstream_retry_cases)) |
| `cacheMaxTime` | string | No | 1m | When the response is selected to be cached in the system, the value of this variable indicates the maximum time to be cached. If cache-control header is not set, the time to be cached will be the defined one. (see [docs](https://github.com/3scale/APIcast/blob/master/doc/parameters.md#apicast_cache_max_time)) |
| `cacheStatusCodes` | string | No | 200 302 | When the response code from upstream matches one of the status codes defined in this environment variable, the response content will be cached (see [docs](https://github.com/3scale/APIcast/blob/master/doc/parameters.md#apicast_cache_status_codes)) |
| `oidcLogLevel` | string | No | err | Allows to set the log level for the logs related to OpenID Connect integration (see [docs](https://github.com/3scale/APIcast/blob/master/doc/parameters.md#apicast_oidc_log_level)) |
| `loadServicesWhenNeeded` | bool | No | false | The configurations are loaded lazily (see [docs](https://github.com/3scale/APIcast/blob/master/doc/parameters.md#apicast_load_services_when_needed)) |
| `servicesFilterByURL` | string | No | N/A |  Used to filter the service configured in the 3scale API Manager, the filter matches with the public base URL (Staging or production) (see [docs](https://github.com/3scale/APIcast/blob/master/doc/parameters.md#apicast_services_filter_by_url)) |
| `serviceConfigurationVersionOverride` | [Service Configuration Version Override object](#service-configuration-version-override-map) | No | N/A | Service configuration version map to prevent it from auto-updating (see [docs](https://github.com/3scale/APIcast/blob/master/doc/parameters.md#apicast_service_id_configuration_version)) |
| `httpsPort` | int | No | 8443 only when `httpsCertificateSecretRef` is provided | Controls on which port APIcast should start listening for HTTPS connections. Do not use `8080` as HTTPS port (see [docs](https://github.com/3scale/APIcast/blob/master/doc/parameters.md#apicast_https_port)) |
| `httpsVerifyDepth` | int | No | N/A | Defines the maximum length of the client certificate chain. (see [docs](https://github.com/3scale/APIcast/blob/master/doc/parameters.md#apicast_https_verify_depth)) |
| `httpsCertificateSecretRef` | LocalObjectReference | No | APIcast has a default certificate used when `httpsPort` is provided | References secret containing the X.509 certificate in the PEM format and the X.509 certificate secret key |
| `caCertificateSecretRef` | LocalObjectReference | No | N/A | References secret containing the X.509 CA certificate |
| `workers` | integer | No | Automatically computed. Check [apicast doc](https://github.com/3scale/APIcast/blob/master/doc/parameters.md#apicast_workers) for further info. | Defines the number of worker processes |
| `timezone` | string | No | N/A | The local timezone of the APIcast deployment pods. Its value must be a compatible value with the tz database | Defines the number of worker processes |
| `customPolicies` | [][CustomPolicySpec](#CustomPolicySpec) | No | N/A | List of custom policies |
| `extendedMetrics` | bool | No | false | Enables additional information on Prometheus metrics (see [docs](https://github.com/3scale/APIcast/blob/master/doc/parameters.md#apicast_extended_metrics)) |
| `customEnvironments` | [][CustomEnvironmentSpec](#CustomEnvironmentSpec) | No | N/A | List of custom environments |
| `openTracing` | [OpenTracingSpec](#OpenTracingSpec) | No | N/A | **[DEPRECATED]** Use `openTelementry` instead. Contains the OpenTracing integration configuration |
| `allProxy` | string | No | N/A | Specifies a HTTP(S) proxy to be used for connecting to services if a protocol-specific proxy is not specified. Authentication is not supported. Format is: `<scheme>://<host>:<port>` (see [docs](https://github.com/3scale/APIcast/blob/master/doc/parameters.md#all_proxy-all_proxy)) |
| `httpProxy` | string | No | N/A | Specifies a HTTP(S) Proxy to be used for connecting to HTTP services. Authentication is not supported. Format is: `<scheme>://<host>:<port>` (see [docs](https://github.com/3scale/APIcast/blob/master/doc/parameters.md#http_proxy-http_proxy)) |
| `httpsProxy` | string | No | N/A | Specifies a HTTP(S) Proxy to be used for connecting to HTTPS services. Authentication is not supported. Format is: `<scheme>://<host>:<port>` (see [docs](https://github.com/3scale/APIcast/blob/master/doc/parameters.md#https_proxy-https_proxy)) |
| `noProxy` | string | No | N/A | Specifies a comma-separated list of hostnames and domain names for which the requests should not be proxied. Setting to a single `*` character, which matches all hosts, effectively disables the proxy (see [docs](https://github.com/3scale/APIcast/blob/master/doc/parameters.md#no_proxy-no_proxy)) |
| `serviceCacheSize` | int | No | N/A | Specifies the number of services that APICast can store in the internal cache (see [docs](https://github.com/3scale/APIcast/blob/master/doc/parameters.md#apicast_service_cache_size)) |
| `openTelemetry` | [OpenTelemetrySpec](#OpenTelemetrySpec) | No | N/A | contains the OpenTelemetry integration configuration |
| `hpa` | bool | No | N/A | When this parameter is set to true, Horizontal Pod Autoscaling will be enabled with default values, spec.replicas and resources limits and requests will be ignored |

#### APIcastStatus

Used by the Operator/Kubernetes to control the state of the Apicast custom resource. It should never be modified by the user.

| **json/yaml field** | **Type** | **Description** |
| --- | --- | --- |
| `image` | string | The image being used in the APIcast deployment |

#### APIcastExposedHost

| **json/yaml field** | **Type** | **Required** | **Default value** | **Description** |
| --- | --- | --- | --- | --- |
| `host` | string | Yes | N/A | Domain name being routed to the gateway |
| `tls` | []networkv1.IngressTLS | No | N/A | Array of ingress TLS objects (see [doc](https://kubernetes.io/docs/concepts/services-networking/ingress/#tls)) |

#### AdminPortalSecret

| **Field** | **Description** |
| --- | --- |
| AdminPortalURL | URI that includes your password and 3scale [Porta](https://github.com/3scale/porta/) portal endpoint. See [format](https://github.com/3scale/APIcast/blob/master/doc/parameters.md#threescale_portal_endpoint) |

#### EmbeddedConfSecret

| **Field** | **Description** |
| --- | --- |
| `config.json` | JSON file with the configuration for the gateway. See [docs](https://github.com/3scale/APIcast/blob/master/doc/parameters.md#threescale_config_file) |

#### Service Configuration Version Override Map

| **Field** | **Value** |
| --- | --- |
| Service `ID` (string type) | The configuration version you can see in the configuration history on the Admin Portal (string type)|

For example, fix service `"2555417833738"` to version `"5"` and service `"2555417836536"` to version `"7"`:

```yaml
spec:
  serviceConfigurationVersionOverride:
    "2555417833738": "5"
    "2555417836536": "7"
```

#### CustomPolicySpec

| **json/yaml field** | **Type** | **Required** | **Default value** | **Description** |
| --- | --- | --- | --- | --- |
| `name` | string | Yes | N/A | Name |
| `version` | string | Yes | N/A | Version |
| `secretRef` | LocalObjectReference | Yes | N/A | Secret reference with the policy content. See [CustomPolicySecret](#CustomPolicySecret) for more information.

#### CustomPolicySecret

Contains custom policy specific content. Two files, `init.lua` and `apicast-policy.json`, are required, but more can be added optionally.

Some examples are available [here](/doc/adding-custom-policies.md)

| **Field** | **Description** |
| --- | --- |
| `init.lua` | Custom policy lua code entry point |
| `apicast-policy.json` | Custom policy metadata |


#### CustomEnvironmentSpec

| **json/yaml field** | **Type** | **Required** | **Default value** | **Description** |
| --- | --- | --- | --- | --- |
| `secretRef` | LocalObjectReference | Yes | N/A | Secret reference with the custom environment content. See [CustomEnvironmentSecret](#CustomEnvironmentSecret) for more information.

#### CustomEnvironmentSecret

Generic (`opaque`) type secret holding one or more custom environments.
The operator will load in the APIcast container all the files (keys) found in the secret.

Some examples are available [here](/doc/adding-custom-environments.md)

| **Field** | **Description** |
| --- | --- |
| *filename* | Custom environment lua code |

### OpenTracingSpec
| **json/yaml field** | **Type** | **Required** | **Default value** | **Description** |
| --- | --- | --- | --- | --- |
| `enabled` | bool | No | `false` | Controls whether OpenTracing integration with APIcast is enabled. By default it is not enabled |
| `tracingLibrary` | string | No | `jaeger` | Controls which OpenTracing library is loaded. At the moment the supported values are: `jaeger`. If not set, `jaeger` will be used |
| `tracingConfigSecretRef` | LocalObjectReference | No | tracing library-specific default | Secret reference with the tracing library-specific configuration. Each supported tracing library provides a default configuration file which is used if `tracingConfigSecretRef` is not specified. See [TracingConfigSecret](#TracingConfigSecret) for more information. |

### TracingConfigSecret

| **Field** | **Description** |
| --- | --- |
| `config` | Tracing library-specific configuration |

**Watch for secret changes**

By default, content changes in the secret will not be noticed by the apicast operator.
The apicast operator allows monitoring the secret for changes adding the `apicast.apps.3scale.net/watched-by=apicast` label.
With that label in place, when the content of the secret is changed, the operator will get notified.
Then, the operator will rollout apicast deployment to make the changes effective.
The operator will not take *ownership* of the secret in any way.

```
kubectl label secret ${SOME_SECRET_NAME} apicast.apps.3scale.net/watched-by=apicast
```

### OpenTelemetrySpec
| **json/yaml field** | **Type** | **Required** | **Default value** | **Description** |
| --- | --- | --- | --- | --- |
| `enabled` | bool | No | `false` | Controls whether opentelemetry based gateway instrumentation is enabled or not. By default it is **disabled** |
| `tracingConfigSecretRef` | *LocalObjectReference* | No | None | Secret reference with the [opentelemetry tracing configuration](https://github.com/open-telemetry/opentelemetry-cpp-contrib/tree/main/instrumentation/nginx). |
| `tracingConfigSecretKey` | string | No | If unspecified, the first secret key in lexicographical order will be referenced as tracing configuration | The secret key used as tracing configuration |

**Watch for secret changes**

By default, content changes in the secret will not be noticed by the apicast operator.
The apicast operator allows monitoring the secret for changes adding the `apicast.apps.3scale.net/watched-by=apicast` label.
With that label in place, when the content of the secret is changed, the operator will get notified.
Then, the operator will rollout apicast deployment to make the changes effective.
The operator will not take *ownership* of the secret in any way.