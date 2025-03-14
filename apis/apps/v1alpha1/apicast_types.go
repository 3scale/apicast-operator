/*
Copyright 2020 Red Hat.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"fmt"

	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"

	appscommon "github.com/3scale/apicast-operator/apis/apps"
	"github.com/3scale/apicast-operator/pkg/k8sutils"
	"github.com/3scale/apicast-operator/version"
)

const (
	APIcastOperatorVersionAnnotation          = "apicast.apps.3scale.net/operator-version"
	APIcastThreescaleVersionAnnotation        = "apicast.apps.3scale.net/apicast-threescale-version"
	ReadyConditionType                 string = "Ready"
	WarningConditionType               string = "Warning"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CustomEnvironmentSpec contains or has reference to an APIcast custom environment
type CustomEnvironmentSpec struct {
	SecretRef *v1.LocalObjectReference `json:"secretRef"`
}

// CustomPolicySpec contains or has reference to an APIcast custom policy
type CustomPolicySpec struct {
	// Name specifies the name of the custom policy
	Name string `json:"name"`
	// Version specifies the name of the custom policy
	Version string `json:"version"`

	// SecretRef specifies the secret holding the custom policy metadata and lua code
	SecretRef *v1.LocalObjectReference `json:"secretRef"`
}

func (c *CustomPolicySpec) VersionName() string {
	return fmt.Sprintf("%s%s", c.Name, c.Version)
}

// APIcastSpec defines the desired state of APIcast.
type APIcastSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Number of replicas of the APIcast Deployment.
	// +optional
	Replicas *int64 `json:"replicas,omitempty"`
	// Enables/disables HPA
	//+optional
	Hpa bool `json:"hpa,omitempty"`
	// Secret reference to a Kubernetes Secret containing the admin portal
	// endpoint URL. The Secret must be located in the same namespace.
	// +optional
	AdminPortalCredentialsRef *v1.LocalObjectReference `json:"adminPortalCredentialsRef,omitempty"`
	// Secret reference to a Kubernetes secret containing the gateway
	// configuration. The Secret must be located in the same namespace.
	// +optional
	EmbeddedConfigurationSecretRef *v1.LocalObjectReference `json:"embeddedConfigurationSecretRef,omitempty"`
	// Kubernetes Service Account name to be used for the APIcast Deployment. The
	// Service Account must exist beforehand.
	// +optional
	ServiceAccount *string `json:"serviceAccount,omitempty"`
	// Image allows overriding the default APIcast gateway container image.
	// This setting should only be used for dev/testing purposes. Setting
	// this disables automated upgrades of the image.
	// +optional
	Image *string `json:"image,omitempty"`
	// ExposedHost is the domain name used for external access. By default no
	// external access is configured.
	// +optional
	ExposedHost *APIcastExposedHost `json:"exposedHost,omitempty"`
	// DeploymentEnvironment is the environment for which the configuration will
	// be downloaded from 3scale (Staging or Production), when using APIcast.
	// The value will also be used in the header X-3scale-User-Agent in the
	// authorize/report requests made to 3scale Service Management API. It is
	// used by 3scale for statistics.
	// +optional
	DeploymentEnvironment *DeploymentEnvironmentType `json:"deploymentEnvironment,omitempty"` // THREESCALE_DEPLOYMENT_ENV
	// DNSResolverAddress can be used to specify a custom DNS resolver address
	// to be used by OpenResty.
	// +optional
	DNSResolverAddress *string `json:"dnsResolverAddress,omitempty"` // RESOLVER
	// EnabledServices can be used to specify a list of service IDs used to
	// filter the configured services.
	// +optional
	EnabledServices []string `json:"enabledServices,omitempty"` // APICAST_SERVICES_LIST
	// ConfigurationLoadMode can be used to set APIcast's configuration load mode.
	// +optional
	// +kubebuilder:validation:Enum=boot;lazy
	ConfigurationLoadMode *string `json:"configurationLoadMode,omitempty"` // APICAST_CONFIGURATION_LOADER
	// LogLevel controls the log level of APIcast's OpenResty logs.
	// +optional
	// +kubebuilder:validation:Enum=debug;info;notice;warn;error;crit;alert;emerg
	LogLevel *string `json:"logLevel,omitempty"` // APICAST_LOG_LEVEL
	// PathRoutingEnabled can be used to enable APIcast's path-based routing
	// in addition to to the default host-based routing.
	// +optional
	PathRoutingEnabled *bool `json:"pathRoutingEnabled,omitempty"` // APICAST_PATH_ROUTING
	// ResponseCodesIncluded can be set to log the response codes of the responses
	// in Apisonator, so they can then be visualized in the 3scale admin portal.
	// +optional
	ResponseCodesIncluded *bool `json:"responseCodesIncluded,omitempty"` // APICAST_RESPONSE_CODES
	// The period (in seconds) that the APIcast configuration will be stored in
	// APIcast's cache.
	// +optional
	CacheConfigurationSeconds *int64 `json:"cacheConfigurationSeconds,omitempty"` // APICAST_CONFIGURATION_CACHE
	// ManagementAPIScope controls APIcast Management API scope. The Management
	// API is powerful and can control the APIcast configuration. debug level
	// should only be enabled for debugging purposes.
	// +optional
	// +kubebuilder:validation:Enum=disabled;status;policies;debug
	ManagementAPIScope *string `json:"managementAPIScope,omitempty"` // APICAST_MANAGEMENT_API
	// OpenSSLPeerVerificationEnabled controls OpenSSL peer verification.
	// +optional
	OpenSSLPeerVerificationEnabled *bool `json:"openSSLPeerVerificationEnabled,omitempty"` // OPENSSL_VERIFY
	// Resources can be used to set custom compute Kubernetes Resource
	// Requirements for the APIcast deployment.
	// +optional
	Resources *v1.ResourceRequirements `json:"resources,omitempty"`
	// UpstreamRetryCases Used only when the retry policy is configured. Specified in which cases a request to the upstream API should be retried.
	// +kubebuilder:validation:Enum=error;timeout;invalid_header;http_500;http_502;http_503;http_504;http_403;http_404;http_429;non_idempotent; off
	// +optional
	UpstreamRetryCases *string `json:"upstreamRetryCases,omitempty"` // APICAST_UPSTREAM_RETRY_CASES
	// CacheMaxTime indicates the maximum time to be cached. If cache-control header is not set, the time to be cached will be the defined one.
	// +optional
	CacheMaxTime *string `json:"cacheMaxTime,omitempty"` // APICAST_CACHE_MAX_TIME
	// CacheStatusCodes defines the status codes for which the response content will be cached.
	// +optional
	CacheStatusCodes *string `json:"cacheStatusCodes,omitempty"` // APICAST_CACHE_STATUS_CODES
	// ServiceCacheSize specifies the number of services that APICast can store in the internal cache
	// +optional
	ServiceCacheSize *int32 `json:"serviceCacheSize,omitempty"`
	// OidcLogLevel allows to set the log level for the logs related to OpenID Connect integration.
	// +kubebuilder:validation:Enum=debug;info;notice;warn;error;crit;alert;emerg
	// +optional
	OidcLogLevel *string `json:"oidcLogLevel,omitempty"` // APICAST_OIDC_LOG_LEVEL
	// LoadServicesWhenNeeded makes the configurations to be loaded lazily. APIcast will only load the ones configured for the host specified in the host header of the request.
	// +optional
	LoadServicesWhenNeeded *bool `json:"loadServicesWhenNeeded,omitempty"` // APICAST_LOAD_SERVICES_WHEN_NEEDED
	// ServicesFilterByURL is used to filter the service configured in the 3scale API Manager, the filter matches with the public base URL (Staging or production).
	// +optional
	ServicesFilterByURL *string `json:"servicesFilterByURL,omitempty"` // APICAST_SERVICES_FILTER_BY_URL
	// ServiceConfigurationVersionOverride contains service configuration version map to prevent it from auto-updating.
	// +optional
	ServiceConfigurationVersionOverride map[string]string `json:"serviceConfigurationVersionOverride,omitempty"` // APICAST_SERVICE_${ID}_CONFIGURATION_VERSION
	// HttpsPort controls on which port APIcast should start listening for HTTPS connections. If this clashes with HTTP port it will be used only for HTTPS.
	// +optional
	HTTPSPort *int32 `json:"httpsPort,omitempty"` // APICAST_HTTPS_PORT
	// HTTPSVerifyDepth defines the maximum length of the client certificate chain.
	// +kubebuilder:validation:Minimum=0
	// +optional
	HTTPSVerifyDepth *int64 `json:"httpsVerifyDepth,omitempty"` // APICAST_HTTPS_VERIFY_DEPTH
	// HTTPSCertificateSecretRef references secret containing the X.509 certificate in the PEM format and the X.509 certificate secret key.
	// +optional
	HTTPSCertificateSecretRef *v1.LocalObjectReference `json:"httpsCertificateSecretRef,omitempty"`
	// CACertificateSecretRef references secret containing the X.509 CA certificate in the PEM format.
	// +optional
	CACertificateSecretRef *v1.LocalObjectReference `json:"caCertificateSecretRef,omitempty"`
	// Workers defines the number of APIcast's worker processes per pod.
	// +optional
	// +kubebuilder:validation:Minimum=1
	Workers *int32 `json:"workers,omitempty"` // APICAST_WORKERS
	// Timezone specifies the local timezone of the APIcast deployment pods. A timezone value available in the TZ database must be set.
	// +optional
	Timezone *string `json:"timezone,omitempty"` // TZ

	// CustomPolicies specifies an array of defined custome policies to be loaded
	// +optional
	CustomPolicies []CustomPolicySpec `json:"customPolicies,omitempty"`

	// ExtendedMetrics enables additional information on Prometheus metrics; some labels will be used with specific information that will provide more in-depth details about APIcast.
	// +optional
	ExtendedMetrics *bool `json:"extendedMetrics,omitempty"` // APICAST_EXTENDED_METRICS

	// CustomEnvironments specifies an array of defined custome environments to be loaded
	// +optional
	CustomEnvironments []CustomEnvironmentSpec `json:"customEnvironments,omitempty"` // APICAST_ENVIRONMENT

	// OpenTracingSpec contains the OpenTracing integration configuration
	// with APIcast.
	// Deprecated
	// +optional
	OpenTracing *OpenTracingSpec `json:"openTracing,omitempty"`

	// AllProxy specifies a HTTP(S) proxy to be used for connecting to services if
	// a protocol-specific proxy is not specified. Authentication is not supported.
	// Format is <scheme>://<host>:<port>
	// +optional
	AllProxy *string `json:"allProxy,omitempty"` // ALL_PROXY

	// HTTPProxy specifies a HTTP(S) Proxy to be used for connecting to HTTP services.
	// Authentication is not supported. Format is <scheme>://<host>:<port>
	// +optional
	HTTPProxy *string `json:"httpProxy,omitempty"` // HTTP_PROXY

	// HTTPSProxy specifies a HTTP(S) Proxy to be used for connecting to HTTPS services.
	// Authentication is not supported. Format is <scheme>://<host>:<port>
	// +optional
	HTTPSProxy *string `json:"httpsProxy,omitempty"` // HTTPS_PROXY

	// NoProxy specifies a comma-separated list of hostnames and domain
	// names for which the requests should not be proxied. Setting to a single
	// * character, which matches all hosts, effectively disables the proxy.
	// +optional
	NoProxy *string `json:"noProxy,omitempty"` // NO_PROXY

	// OpenTelemetry contains the gateway instrumentation configuration
	// with APIcast.
	// +optional
	OpenTelemetry *OpenTelemetrySpec `json:"openTelemetry,omitempty"`
}

func (a *APIcast) OpenTracingIsEnabled() bool {
	return a.Spec.OpenTracing != nil && a.Spec.OpenTracing.Enabled != nil && *a.Spec.OpenTracing.Enabled
}

func (a *APIcast) OpenTelemetryEnabled() bool {
	return a.Spec.OpenTelemetry != nil && a.Spec.OpenTelemetry.Enabled != nil && *a.Spec.OpenTelemetry.Enabled
}

type DeploymentEnvironmentType string

const (
	DefaultHTTPPort  int32 = 8080
	DefaultHTTPSPort int32 = 8443
)

type APIcastExposedHost struct {
	Host string `json:"host"`
	// +optional
	TLS []networkingv1.IngressTLS `json:"tls,omitempty"`
}

type OpenTelemetrySpec struct {
	// Enabled controls whether OpenTelemetry integration with APIcast is enabled.
	// By default it is not enabled.
	// +optional
	Enabled *bool `json:"enabled,omitempty"`

	// TracingConfigSecretRef contains a Secret reference the Opentelemetry configuration.
	// The configuration file specification is defined in the Nginx instrumentation library repo
	// https://github.com/open-telemetry/opentelemetry-cpp-contrib/tree/main/instrumentation/nginx
	// +optional
	TracingConfigSecretRef *v1.LocalObjectReference `json:"tracingConfigSecretRef,omitempty"`

	// TracingConfigSecretKey contains the key of the secret to select the configuration from.
	// if unspecified, the first secret key in lexicographical order will be selected.
	// +optional
	TracingConfigSecretKey *string `json:"tracingConfigSecretKey,omitempty"`
}

type OpenTracingSpec struct {
	// Enabled controls whether OpenTracing integration with APIcast is enabled.
	// By default it is not enabled.
	// +optional
	Enabled *bool `json:"enabled,omitempty"`
	// TracingLibrary controls which OpenTracing library is loaded. At the moment
	// the only supported tracer is `jaeger`. If not set, `jaeger` will be used.
	// +optional
	TracingLibrary *string `json:"tracingLibrary,omitempty"`
	// TracingConfigSecretRef contains a Secret reference the OpenTracing configuration.
	// Each supported tracing library provides a default configuration file
	// that is used if TracingConfig is not specified.
	// +optional
	TracingConfigSecretRef *v1.LocalObjectReference `json:"tracingConfigSecretRef,omitempty"`
}

// APIcastStatus defines the observed state of APIcast.
type APIcastStatus struct {
	// The image being used in the APIcast deployment.
	// +optional
	Image string `json:"image,omitempty"`

	// ObservedGeneration reflects the generation of the most recently observed spec.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Represents the observations of a foo's current state.
	// Known .status.conditions.type are: "Available"
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`
}

func (r *APIcastStatus) IsReady() bool {
	for i := range r.Conditions {
		if r.Conditions[i].Type == ReadyConditionType {
			return r.Conditions[i].Status == metav1.ConditionTrue
		}
	}

	return false
}

func (r *APIcastStatus) Equals(other *APIcastStatus, logger logr.Logger) bool {
	if r.Image != other.Image {
		diff := cmp.Diff(r.Image, other.Image)
		logger.V(1).Info("Image not equal", "difference", diff)
		return false
	}

	if r.ObservedGeneration != other.ObservedGeneration {
		diff := cmp.Diff(r.ObservedGeneration, other.ObservedGeneration)
		logger.V(1).Info("ObservedGeneration not equal", "difference", diff)
		return false
	}

	// Marshalling sorts by condition type
	currentMarshaledJSON, _ := k8sutils.ConditionMarshal(r.Conditions)
	otherMarshaledJSON, _ := k8sutils.ConditionMarshal(other.Conditions)
	if string(currentMarshaledJSON) != string(otherMarshaledJSON) {
		diff := cmp.Diff(string(currentMarshaledJSON), string(otherMarshaledJSON))
		logger.V(1).Info("Conditions not equal", "difference", diff)
		return false
	}

	return true
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// APIcast is the Schema for the apicasts API.
// +kubebuilder:resource:path=apicasts,scope=Namespaced
// +operator-sdk:csv:customresourcedefinitions:displayName="APIcast"
type APIcast struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   APIcastSpec   `json:"spec,omitempty"`
	Status APIcastStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// APIcastList contains a list of APIcasts.
type APIcastList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []APIcast `json:"items"`
}

func (a *APIcast) GetOwnerReference() *metav1.OwnerReference {
	trueVar := true
	return &metav1.OwnerReference{
		APIVersion: GroupVersion.String(),
		Kind:       appscommon.APIcastKind,
		Name:       a.Name,
		UID:        a.UID,
		Controller: &trueVar,
	}
}

func (a *APIcast) UpdateOperatorVersion() bool {
	changed := false

	if a.Annotations == nil {
		a.Annotations = map[string]string{}
		changed = true
	}

	if v, ok := a.Annotations[APIcastOperatorVersionAnnotation]; !ok || v != version.Version {
		a.Annotations[APIcastOperatorVersionAnnotation] = version.Version
		changed = true
	}

	if v, ok := a.Annotations[APIcastThreescaleVersionAnnotation]; !ok || v != version.ThreescaleVersionMajorMinorPatch() {
		a.Annotations[APIcastThreescaleVersionAnnotation] = version.ThreescaleVersionMajorMinorPatch()
		changed = true
	}

	return changed
}

func (a *APIcast) Reset() { *a = APIcast{} }

func (a *APIcast) Validate() field.ErrorList {
	errors := field.ErrorList{}

	// check HTTPSPort does not conflict with default HTTPPort
	specFldPath := field.NewPath("spec")
	httpsPortFldPath := specFldPath.Child("httpsPort")

	if a.Spec.HTTPSPort != nil && *a.Spec.HTTPSPort == DefaultHTTPPort {
		errors = append(errors, field.Invalid(httpsPortFldPath, a.Spec.HTTPSPort, "HTTPS port conflicts with HTTP port"))
	}

	customPoliciesFldPath := specFldPath.Child("customPolicies")
	// check custom policy secret is set
	for idx, customPolicySpec := range a.Spec.CustomPolicies {
		if customPolicySpec.SecretRef == nil {
			customPoliciesIdxFldPath := customPoliciesFldPath.Index(idx)
			errors = append(errors, field.Invalid(customPoliciesIdxFldPath, customPolicySpec, "custom policy secret is mandatory"))
		} else if customPolicySpec.SecretRef.Name == "" {
			customPoliciesIdxFldPath := customPoliciesFldPath.Index(idx)
			errors = append(errors, field.Invalid(customPoliciesIdxFldPath, customPolicySpec, "custom policy secret name is empty"))
		}
	}

	// check duplicated custom policy version name
	duplicateMap := make(map[string]int)
	for idx, customPolicySpec := range a.Spec.CustomPolicies {
		if _, ok := duplicateMap[customPolicySpec.VersionName()]; ok {
			customPoliciesIdxFldPath := customPoliciesFldPath.Index(idx)
			errors = append(errors, field.Invalid(customPoliciesIdxFldPath, customPolicySpec, "custom policy secret name version tuple is duplicated"))
			break
		}
		duplicateMap[customPolicySpec.VersionName()] = 0
	}

	customEnvsFldPath := specFldPath.Child("customEnvironments")
	// check custom environment secret is set
	for idx, customEnvSpec := range a.Spec.CustomEnvironments {
		if customEnvSpec.SecretRef == nil {
			customEnvsIdxFldPath := customEnvsFldPath.Index(idx)
			errors = append(errors, field.Invalid(customEnvsIdxFldPath, customEnvSpec, "custom environment secret is mandatory"))
		} else if customEnvSpec.SecretRef.Name == "" {
			customEnvsIdxFldPath := customEnvsFldPath.Index(idx)
			errors = append(errors, field.Invalid(customEnvsIdxFldPath, customEnvSpec, "custom environment secret name is empty"))
		}
	}

	// check tracing config secret has a name specified when tracing config is
	// enabled and a custom configuration secret reference has been set
	if a.OpenTracingIsEnabled() {
		if a.Spec.OpenTracing.TracingConfigSecretRef != nil {
			if a.Spec.OpenTracing.TracingConfigSecretRef.Name == "" {
				openTracingFldPath := specFldPath.Child("openTracing")
				customTracingConfigFldPath := openTracingFldPath.Child("tracingConfigSecretRef")
				errors = append(errors, field.Invalid(customTracingConfigFldPath, a.Spec.OpenTracing, "custom tracing library secret name is empty"))
			}
		}
	}

	// check opentracing config secret has a name specified when tracing config is
	// enabled and a custom configuration secret reference has been set
	if a.OpenTelemetryEnabled() {
		if a.Spec.OpenTelemetry.TracingConfigSecretRef != nil &&
			a.Spec.OpenTelemetry.TracingConfigSecretRef.Name == "" {
			openTracingFldPath := specFldPath.Child("openTelemetry")
			customTracingConfigFldPath := openTracingFldPath.Child("tracingConfigSecretRef")
			errors = append(errors, field.Invalid(customTracingConfigFldPath, a.Spec.OpenTelemetry, "tracing config secret name is empty"))
		}
	}

	return errors
}

func init() {
	SchemeBuilder.Register(&APIcast{}, &APIcastList{})
}

func (a *APIcast) GetAdminPortalCredentialsSecretRef() *v1.LocalObjectReference {
	return a.Spec.AdminPortalCredentialsRef
}

func (a *APIcast) GetEmbeddedConfigurationSecretRef() *v1.LocalObjectReference {
	return a.Spec.EmbeddedConfigurationSecretRef
}

func (a *APIcast) GetOpenTelemetrySecretRef() *v1.LocalObjectReference {
	if a.Spec.OpenTelemetry == nil {
		return nil
	}
	return a.Spec.OpenTelemetry.TracingConfigSecretRef
}

func (a *APIcast) GetOpenTracingSecretRef() *v1.LocalObjectReference {
	if a.Spec.OpenTracing == nil {
		return nil
	}
	return a.Spec.OpenTracing.TracingConfigSecretRef
}

func (a *APIcast) GetCustomEnvironmentsSecretRefs() []*v1.LocalObjectReference {
	if a.Spec.CustomEnvironments == nil {
		return nil
	}

	secretRefs := []*v1.LocalObjectReference{}
	for _, env := range a.Spec.CustomEnvironments {
		if env.SecretRef != nil {
			secretRefs = append(secretRefs, env.SecretRef)
		}
	}

	return secretRefs
}

func (a *APIcast) GetCustomPoliciesSecretRefs() []*v1.LocalObjectReference {
	if a.Spec.CustomPolicies == nil {
		return nil
	}

	secretRefs := []*v1.LocalObjectReference{}
	for _, policy := range a.Spec.CustomPolicies {
		if policy.SecretRef != nil {
			secretRefs = append(secretRefs, policy.SecretRef)
		}
	}

	return secretRefs
}

func (a *APIcast) GetApicastSecretRefs() []*v1.LocalObjectReference {
	secretRefs := []*v1.LocalObjectReference{}

	adminPortalCredentialsSecretRef := a.GetAdminPortalCredentialsSecretRef()
	if adminPortalCredentialsSecretRef != nil {
		secretRefs = append(secretRefs, adminPortalCredentialsSecretRef)
	}

	embeddedConfigurationSecretRef := a.GetEmbeddedConfigurationSecretRef()
	if embeddedConfigurationSecretRef != nil {
		secretRefs = append(secretRefs, embeddedConfigurationSecretRef)
	}

	openTelemetrySecretRef := a.GetOpenTelemetrySecretRef()
	if openTelemetrySecretRef != nil {
		secretRefs = append(secretRefs, openTelemetrySecretRef)
	}

	opentracingSecretRef := a.GetOpenTracingSecretRef()
	if opentracingSecretRef != nil {
		secretRefs = append(secretRefs, opentracingSecretRef)
	}

	customEnvironmentSecretRefs := a.GetCustomEnvironmentsSecretRefs()
	if len(customEnvironmentSecretRefs) > 0 {
		secretRefs = append(secretRefs, customEnvironmentSecretRefs...)
	}

	customPoliciesSecretRefs := a.GetCustomPoliciesSecretRefs()
	if len(customPoliciesSecretRefs) > 0 {
		secretRefs = append(secretRefs, customPoliciesSecretRefs...)
	}

	return secretRefs
}
