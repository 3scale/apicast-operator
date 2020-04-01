package apicast

import (
	validator "github.com/go-playground/validator/v10"
	v1 "k8s.io/api/core/v1"
	extensions "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ExposedHost struct {
	Host string
	TLS  []extensions.IngressTLS
}

type APIcastOptions struct {
	Namespace                    string        `validate:"required"`
	DeploymentName               string        `validate:"required"`
	Owner                        metav1.Object `validate:"required"`
	ServiceName                  string        `validate:"required"`
	Replicas                     int32
	AppLabel                     string            `validate:"required"`
	AdditionalAnnotations        map[string]string `validate:"required"`
	ServiceAccountName           string            `validate:"required"`
	Image                        string            `validate:"required"`
	ExposedHost                  ExposedHost       `validate:"-"`
	AdminPortalCredentialsSecret *v1.Secret        `validate:"required_without=GatewayConfigurationSecret"`
	GatewayConfigurationSecret   *v1.Secret        `validate:"required_without=AdminPortalCredentialsSecret"`

	DeploymentEnvironment          *string
	DNSResolverAddress             *string
	EnabledServices                []string
	ConfigurationLoadMode          *string
	LogLevel                       *string
	PathRoutingEnabled             *bool
	ResponseCodesIncluded          *bool
	CacheConfigurationSeconds      *int64
	ManagementAPIScope             *string
	OpenSSLPeerVerificationEnabled *bool
}

func NewAPIcastOptions() *APIcastOptions {
	return &APIcastOptions{}
}

func (a *APIcastOptions) Validate() error {
	validate := validator.New()
	return validate.Struct(a)
}
