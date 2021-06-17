package apicast

import (
	validator "github.com/go-playground/validator/v10"
	v1 "k8s.io/api/core/v1"
	extensions "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ExposedHost struct {
	Host string
	TLS  []extensions.IngressTLS
}

type CustomPolicy struct {
	Name      string
	Version   string
	SecretRef v1.LocalObjectReference
}

type APIcastOptions struct {
	Namespace                    string                 `validate:"required"`
	DeploymentName               string                 `validate:"required"`
	Owner                        *metav1.OwnerReference `validate:"required"`
	ServiceName                  string                 `validate:"required"`
	Replicas                     int32
	AppLabel                     string                  `validate:"required"`
	AdditionalAnnotations        map[string]string       `validate:"required"`
	ServiceAccountName           string                  `validate:"required"`
	Image                        string                  `validate:"required"`
	ExposedHost                  ExposedHost             `validate:"-"`
	AdminPortalCredentialsSecret *v1.Secret              `validate:"required_without=GatewayConfigurationSecret"`
	GatewayConfigurationSecret   *v1.Secret              `validate:"required_without=AdminPortalCredentialsSecret"`
	ResourceRequirements         v1.ResourceRequirements `validate:"-"`

	DeploymentEnvironment               *string
	DNSResolverAddress                  *string
	EnabledServices                     []string
	ConfigurationLoadMode               *string
	LogLevel                            *string
	PathRoutingEnabled                  *bool
	ResponseCodesIncluded               *bool
	CacheConfigurationSeconds           *int64
	ManagementAPIScope                  *string
	OpenSSLPeerVerificationEnabled      *bool
	UpstreamRetryCases                  *string
	CacheMaxTime                        *string
	CacheStatusCodes                    *string
	OidcLogLevel                        *string
	LoadServicesWhenNeeded              *bool
	ServicesFilterByURL                 *string
	ServiceConfigurationVersionOverride map[string]string
	HTTPSPort                           *int32
	HTTPSVerifyDepth                    *int64
	HTTPSCertificateSecret              *v1.Secret
	Workers                             *int32
	Timezone                            *string
	CustomPolicies                      []CustomPolicy
}

func NewAPIcastOptions() *APIcastOptions {
	return &APIcastOptions{}
}

func (a *APIcastOptions) Validate() error {
	validate := validator.New()
	return validate.Struct(a)
}

func DefaultResourceRequirements() v1.ResourceRequirements {
	return v1.ResourceRequirements{
		Limits: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("1"),
			v1.ResourceMemory: resource.MustParse("128Mi"),
		},
		Requests: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("500m"),
			v1.ResourceMemory: resource.MustParse("64Mi"),
		},
	}
}
