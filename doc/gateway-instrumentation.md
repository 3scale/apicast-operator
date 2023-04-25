## Gateway instrumentation

APIcast gateway can be instrumented using the [OpenTelemetry](https://opentelemetry.io/) SDK.
More specifically, enabling the [Nginx opentelemetry tracing library](https://github.com/open-telemetry/opentelemetry-cpp-contrib/tree/main/instrumentation/nginx).

It works with Jaeger since version **1.35**.  If the existing collector does not support
OpenTelemetry traces, an OpenTelemetry Collector is required as tracing proxy.

Supported propagation types: [W3C](https://w3c.github.io/trace-context/)

### Prerequisites

* Opentelemetry Collector supporting the APIcast exporter. Currently, the only implemeneted [exporter](https://opentelemetry.io/docs/reference/specification/protocol/exporter/)
in APIcast is OTLP over gRPC `OTLP/gRPC`. Even though OpenTelemetry specification supports also OTLP over HTTP (`OTLP/HTTP`),
this exporter is not included in APIcast. It works with Jaeger since version **1.35**.

For dev/testing purposes, you can deploy quick&easy Jaeger with few commands. **Not suitable** for production use, though.

```
❯ make kind-create-cluster
❯ make jaeger-deploy
```

That should deploy Jaeger service listening at `jaeger:4317`.

### Create secret with the APIcast instrumentation configuration

The configuration file specification is defined in the [Nginx instrumentation library repo](https://github.com/open-telemetry/opentelemetry-cpp-contrib/tree/main/instrumentation/nginx).

`otlp` is the only supported exporter.

The host/port address is set according to the dev/testing deployment defined in this repo.
Change it to whatever you have your collector deployed.

The name of the secret. `otel-config` in the example, will be used in the APIcast CR.

```yaml
kubectl apply -f - <<EOF
---
apiVersion: v1
kind: Secret
metadata:
  name: otel-config
type: Opaque
stringData:
  config.json: |
    exporter = "otlp"
    processor = "simple"
    [exporters.otlp]
    # Alternatively the OTEL_EXPORTER_OTLP_ENDPOINT environment variable can also be used.
    host = "jaeger"
    port = 4317
    # Optional: enable SSL, for endpoints that support it
    # use_ssl = true
    # Optional: set a filesystem path to a pem file to be used for SSL encryption
    # (when use_ssl = true)
    # ssl_cert_path = "/path/to/cert.pem"
    [processors.batch]
    max_queue_size = 2048
    schedule_delay_millis = 5000
    max_export_batch_size = 512
    [service]
    name = "apicast" # Opentelemetry resource name
EOF
```

### Deploy APIcast with opentelemetry instrumentation

Only relevant content shown. Check out the [APIcast CRD reference](apicast-crd-reference.md) for
a comprehensive list of options.

```yaml
kubectl apply -f - <<EOF
---
apiVersion: apps.3scale.net/v1alpha1
kind: APIcast
metadata:
  name: apicast1
spec:
  openTelemetry:
    enabled: true
    tracingConfigSecretRef:
      name: otel-config
EOF
```

For dev/testing purposes,
you can deploy quick&easy APIcast with opentelemetry instrumentation enabled with few commands.
**Not suitable** for production use, though.

The APIcast operator must be up&running.

```
❯ make apicast-opentelemetry-deploy
```

### Verification steps for the opentelemetry instrumentation

Testing APIcast instrumentation in the dev/testing environment should be easy.

* Send valid request to APIcast

```
❯ kubectl port-forward service/apicast-apicast1 18080:8080&


❯ curl -v -H "Host: one"  http://127.0.0.1:18080/get?user_key=foo
*   Trying 127.0.0.1:18080...
* Connected to 127.0.0.1 (127.0.0.1) port 18080 (#0)
> GET /get?user_key=foo HTTP/1.1
> Host: one
> User-Agent: curl/7.81.0
> Accept: */*
>
Handling connection for 18080
* Mark bundle as not supporting multiuse
< HTTP/1.1 200 OK
< Server: openresty
< Date: Tue, 25 Apr 2023 10:29:18 GMT
< Content-Type: application/json
< Content-Length: 371
< Connection: keep-alive
< Access-Control-Allow-Origin: *
< Access-Control-Allow-Credentials: true
<
{
  "args": {
    "user_key": "foo"
  },
  "headers": {
    "Accept": "*/*",
    "Host": "httpbin.org",
    "Traceparent": "00-0e5a02ef5281362baa542441fdb0a939-53351ffcdcd80a61-01",
    "User-Agent": "curl/7.81.0",
    "X-Amzn-Trace-Id": "Root=1-6447ab7a-1e08868b059f43631e04a03a"
  },
  "origin": "147.161.83.37",
  "url": "http://httpbin.org/get?user_key=foo"
}
* Connection #0 to host 127.0.0.1 left intact
```

Note that upstream echo'ed request headers show `Traceparent` W3C standard tracing header.

Open in local browser jaeger dashboard

```
❯ kubectl port-forward service/jaeger 16686&

❯ open http://127.0.0.1:16686
```

Hit "Find Traces" with `Service` set to `apicast`. There should be at lease one trace.
