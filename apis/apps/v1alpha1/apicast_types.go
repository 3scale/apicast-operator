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
	appscommon "github.com/3scale/apicast-operator/apis/apps"

	v1 "k8s.io/api/core/v1"
	extensions "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// APIcastSpec defines the desired state of APIcast
type APIcastSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// +optional
	Replicas *int64 `json:"replicas,omitempty"`
	// +optional
	AdminPortalCredentialsRef *v1.LocalObjectReference `json:"adminPortalCredentialsRef,omitempty"`
	// +optional
	EmbeddedConfigurationSecretRef *v1.LocalObjectReference `json:"embeddedConfigurationSecretRef,omitempty"`
	// +optional
	ServiceAccount *string `json:"serviceAccount,omitempty"`
	// +optional
	Image *string `json:"image,omitempty"`
	// +optional
	ExposedHost *APIcastExposedHost `json:"exposedHost,omitempty"`
	// +optional
	DeploymentEnvironment *DeploymentEnvironmentType `json:"deploymentEnvironment,omitempty"` // THREESCALE_DEPLOYMENT_ENV
	// +optional
	DNSResolverAddress *string `json:"dnsResolverAddress,omitempty"` // RESOLVER
	// +optional
	EnabledServices []string `json:"enabledServices,omitempty"` // APICAST_SERVICES_LIST
	// +optional
	// +kubebuilder:validation:Enum=boot;lazy
	ConfigurationLoadMode *string `json:"configurationLoadMode,omitempty"` // APICAST_CONFIGURATION_LOADER
	// +optional
	// +kubebuilder:validation:Enum=debug;info;notice;warn;error;crit;alert;emerg
	LogLevel *string `json:"logLevel,omitempty"` // APICAST_LOG_LEVEL
	// +optional
	PathRoutingEnabled *bool `json:"pathRoutingEnabled,omitempty"` // APICAST_PATH_ROUTING
	// +optional
	ResponseCodesIncluded *bool `json:"responseCodesIncluded,omitempty"` // APICAST_RESPONSE_CODES
	// +optional
	CacheConfigurationSeconds *int64 `json:"cacheConfigurationSeconds,omitempty"` // APICAST_CONFIGURATION_CACHE
	// +optional
	// +kubebuilder:validation:Enum=disabled;status;policies;debug
	ManagementAPIScope *string `json:"managementAPIScope,omitempty"` // APICAST_MANAGEMENT_API
	// +optional
	OpenSSLPeerVerificationEnabled *bool `json:"openSSLPeerVerificationEnabled,omitempty"` // OPENSSL_VERIFY
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
}

type DeploymentEnvironmentType string

const (
	DeploymentEnvironmentProduction = "production"
	DeploymentEnvironmentStaging    = "staging"
)

type APIcastExposedHost struct {
	Host string `json:"host"`
	// +optional
	TLS []extensions.IngressTLS `json:"tls,omitempty"`
}

// APIcastStatus defines the observed state of APIcast
type APIcastStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Represents the latest available observations of a replica set's current state.
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	Conditions []APIcastCondition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`

	// The image being used in the APIcast deployment
	// +optional
	Image string `json:"image,omitempty"`
}

type APIcastConditionType string

type APIcastCondition struct {
	// Type of replica set condition.
	Type APIcastConditionType `json:"type"`
	// Status of the condition, one of True, False, Unknown.
	Status v1.ConditionStatus `json:"status"`

	// The Reason, Message, LastHeartbeatTime and LastTransitionTime fields are
	// optional. Unless we really use them they should directly not be used even
	// if they are optional
	// The last time the condition transitioned from one status to another.
	// +optional
	//LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
	// The reason for the condition's last transition.
	// +optional
	//Reason string `json:"reason,omitempty"`
	// A human readable message indicating details about the transition.
	// +optional
	//Message string `json:"message,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// APIcast is the Schema for the apicasts API
// +kubebuilder:resource:path=apicasts,scope=Namespaced
// +operator-sdk:csv:customresourcedefinitions:displayName="APIcast"
type APIcast struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   APIcastSpec   `json:"spec,omitempty"`
	Status APIcastStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// APIcastList contains a list of APIcast
type APIcastList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []APIcast `json:"items"`
}

func (a *APIcast) GetOwnerRefence() *metav1.OwnerReference {
	trueVar := true
	return &metav1.OwnerReference{
		APIVersion: GroupVersion.String(),
		Kind:       appscommon.APIcastKind,
		Name:       a.Name,
		UID:        a.UID,
		Controller: &trueVar,
	}
}

func (a *APIcast) Reset() { *a = APIcast{} }

func (a *APIcast) Validate() field.ErrorList {
	errors := field.ErrorList{}

	// check HTTPSPort is required for httpsCertificateSecretRef
	// check httpsCertificateSecretRef is required for HTTPSPort
	specFldPath := field.NewPath("spec")
	httpsPortFldPath := specFldPath.Child("httpsPort")
	httpsCertificateSecretRefFldPath := specFldPath.Child("httpsCertificateSecretRef")

	if a.Spec.HTTPSPort != nil && a.Spec.HTTPSCertificateSecretRef == nil {
		errors = append(errors, field.Required(httpsCertificateSecretRefFldPath, "credentials secret is required when https port is set"))
	}

	if a.Spec.HTTPSCertificateSecretRef != nil && a.Spec.HTTPSPort == nil {
		errors = append(errors, field.Required(httpsPortFldPath, "https port is required when credentials secret is provided"))
	}

	return errors
}

func init() {
	SchemeBuilder.Register(&APIcast{}, &APIcastList{})
}
