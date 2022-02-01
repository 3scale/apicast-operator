package apicast

import (
	validator "github.com/go-playground/validator/v10"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ExposedHost struct {
	Host string
	TLS  []networkingv1.IngressTLS
}

type CustomPolicy struct {
	Name    string
	Version string
	Secret  *v1.Secret
}

type TracingConfig struct {
	Enabled        bool
	TracingLibrary string `validate:"required"`
	Secret         *v1.Secret
}

type APIcastOptions struct {
	Namespace                    string                 `validate:"required"`
	DeploymentName               string                 `validate:"required"`
	Owner                        *metav1.OwnerReference `validate:"required"`
	ServiceName                  string                 `validate:"required"`
	Replicas                     int32
	AdditionalPodAnnotations     map[string]string       `validate:"required"`
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
	ExtendedMetrics                     *bool
	CustomEnvironments                  []*v1.Secret
	TracingConfig                       TracingConfig `validate:"-"`
	AllProxy                            *string
	HTTPProxy                           *string
	HTTPSProxy                          *string
	NoProxy                             *string

	CommonLabels      map[string]string `validate:"required"`
	PodTemplateLabels map[string]string `validate:"required"`
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
