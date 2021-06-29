package apicast

import (
	"context"
	"fmt"
	"net/url"

	appsv1alpha1 "github.com/3scale/apicast-operator/apis/apps/v1alpha1"
	"github.com/3scale/apicast-operator/pkg/k8sutils"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	AdmPortalSecretResverAnnotation            = "apicast.apps.3scale.net/admin-portal-secret-resource-version"
	GatewayConfigurationSecretResverAnnotation = "apicast.apps.3scale.net/gateway-configuration-secret-resource-version"
	APPLABEL                                   = "apicast"
)

type APIcastOptionsProvider struct {
	APIcastCR      *appsv1alpha1.APIcast
	APIcastOptions *APIcastOptions
	Client         client.Client
}

func NewApicastOptionsProvider(cr *appsv1alpha1.APIcast, cl client.Client) *APIcastOptionsProvider {
	return &APIcastOptionsProvider{
		APIcastCR:      cr,
		APIcastOptions: NewAPIcastOptions(),
		Client:         cl,
	}
}

func (a *APIcastOptionsProvider) GetApicastOptions() (*APIcastOptions, error) {
	a.APIcastOptions.Namespace = a.APIcastCR.Namespace
	a.APIcastOptions.Owner = a.APIcastCR.GetOwnerRefence()

	apicastFullName := "apicast-" + a.APIcastCR.Name
	a.APIcastOptions.DeploymentName = apicastFullName
	a.APIcastOptions.ServiceName = apicastFullName
	a.APIcastOptions.Replicas = int32(*a.APIcastCR.Spec.Replicas)
	a.APIcastOptions.AppLabel = APPLABEL

	a.APIcastOptions.ServiceAccountName = "default"
	if a.APIcastCR.Spec.ServiceAccount != nil {
		a.APIcastOptions.ServiceAccountName = *a.APIcastCR.Spec.ServiceAccount
	}

	a.APIcastOptions.Image = GetDefaultImageVersion()
	if a.APIcastCR.Spec.Image != nil {
		a.APIcastOptions.Image = *a.APIcastCR.Spec.Image
	}

	a.APIcastOptions.ExposedHost = ExposedHost{}
	if a.APIcastCR.Spec.ExposedHost != nil {
		a.APIcastOptions.ExposedHost.Host = a.APIcastCR.Spec.ExposedHost.Host
		a.APIcastOptions.ExposedHost.TLS = a.APIcastCR.Spec.ExposedHost.TLS
	}

	adminPortalCredentialsSecret, err := a.getAdminPortalCredentialsSecret()
	if err != nil {
		return nil, err
	}
	a.APIcastOptions.AdminPortalCredentialsSecret = adminPortalCredentialsSecret

	gatewayConfigurationSecret, err := a.getGatewayEmbeddedConfigSecret()
	if err != nil {
		return nil, err
	}
	a.APIcastOptions.GatewayConfigurationSecret = gatewayConfigurationSecret

	if a.APIcastCR.Spec.DeploymentEnvironment != nil {
		res := string(*a.APIcastCR.Spec.DeploymentEnvironment)
		a.APIcastOptions.DeploymentEnvironment = &res
	}

	a.APIcastOptions.DNSResolverAddress = a.APIcastCR.Spec.DNSResolverAddress
	a.APIcastOptions.EnabledServices = a.APIcastCR.Spec.EnabledServices
	a.APIcastOptions.ConfigurationLoadMode = a.APIcastCR.Spec.ConfigurationLoadMode
	a.APIcastOptions.LogLevel = a.APIcastCR.Spec.LogLevel
	a.APIcastOptions.PathRoutingEnabled = a.APIcastCR.Spec.PathRoutingEnabled
	a.APIcastOptions.ResponseCodesIncluded = a.APIcastCR.Spec.ResponseCodesIncluded
	a.APIcastOptions.CacheConfigurationSeconds = a.APIcastCR.Spec.CacheConfigurationSeconds
	a.APIcastOptions.ManagementAPIScope = a.APIcastCR.Spec.ManagementAPIScope
	a.APIcastOptions.OpenSSLPeerVerificationEnabled = a.APIcastCR.Spec.OpenSSLPeerVerificationEnabled
	a.APIcastOptions.UpstreamRetryCases = a.APIcastCR.Spec.UpstreamRetryCases
	a.APIcastOptions.CacheMaxTime = a.APIcastCR.Spec.CacheMaxTime
	a.APIcastOptions.CacheStatusCodes = a.APIcastCR.Spec.CacheStatusCodes
	a.APIcastOptions.OidcLogLevel = a.APIcastCR.Spec.OidcLogLevel
	a.APIcastOptions.LoadServicesWhenNeeded = a.APIcastCR.Spec.LoadServicesWhenNeeded
	a.APIcastOptions.ServicesFilterByURL = a.APIcastCR.Spec.ServicesFilterByURL
	a.APIcastOptions.ServiceConfigurationVersionOverride = a.APIcastCR.Spec.ServiceConfigurationVersionOverride
	a.APIcastOptions.HTTPSPort = a.APIcastCR.Spec.HTTPSPort
	// when HTTPS certificate is provided and HTTPS port is not provided, assing default https port
	if a.APIcastCR.Spec.HTTPSCertificateSecretRef != nil && a.APIcastCR.Spec.HTTPSPort == nil {
		tmpDefaultPort := appsv1alpha1.DefaultHTTPSPort
		a.APIcastOptions.HTTPSPort = &tmpDefaultPort
	}
	a.APIcastOptions.HTTPSVerifyDepth = a.APIcastCR.Spec.HTTPSVerifyDepth

	// when HTTPS port is provided and HTTPS Certificate secret is not provided,
	// Apicast will use some default certificate
	// Should the operator raise a warning?
	httpsCertificateSecret, err := a.getHTTPSCertificateSecret()
	if err != nil {
		return nil, err
	}
	a.APIcastOptions.HTTPSCertificateSecret = httpsCertificateSecret

	// Annotations from user secrets. Used to rollout apicast deployment if any secrets changes
	a.APIcastOptions.AdditionalAnnotations = a.additionalAnnotations()

	// Resource requirements
	resourceRequirements := DefaultResourceRequirements()
	if a.APIcastCR.Spec.Resources != nil {
		resourceRequirements = *a.APIcastCR.Spec.Resources
	}
	a.APIcastOptions.ResourceRequirements = resourceRequirements

	a.APIcastOptions.Workers = a.APIcastCR.Spec.Workers
	a.APIcastOptions.Timezone = a.APIcastCR.Spec.Timezone

	for idx, customPolicySpec := range a.APIcastCR.Spec.CustomPolicies {
		namespacedName := types.NamespacedName{
			Name:      customPolicySpec.SecretRef.Name, // CR Validation ensures not nil
			Namespace: a.APIcastCR.Namespace,
		}
		err := a.validateCustomPolicySecret(namespacedName)
		if err != nil {
			errors := field.ErrorList{}
			customPoliciesIdxFldPath := field.NewPath("spec").Child("customPolicies").Index(idx)
			errors = append(errors, field.Invalid(customPoliciesIdxFldPath, customPolicySpec, err.Error()))
			return nil, errors.ToAggregate()
		}

		a.APIcastOptions.CustomPolicies = append(a.APIcastOptions.CustomPolicies, CustomPolicy{
			Name:      customPolicySpec.Name,
			Version:   customPolicySpec.Version,
			SecretRef: *customPolicySpec.SecretRef,
		})
	}

	a.APIcastOptions.ExtendedMetrics = a.APIcastCR.Spec.ExtendedMetrics

	return a.APIcastOptions, a.APIcastOptions.Validate()
}

func (a *APIcastOptionsProvider) additionalAnnotations() map[string]string {
	annotations := map[string]string{}

	if a.APIcastOptions.AdminPortalCredentialsSecret != nil {
		annotations[AdmPortalSecretResverAnnotation] = a.APIcastOptions.AdminPortalCredentialsSecret.ResourceVersion
	}

	if a.APIcastOptions.GatewayConfigurationSecret != nil {
		annotations[GatewayConfigurationSecretResverAnnotation] = a.APIcastOptions.GatewayConfigurationSecret.ResourceVersion
	}

	return annotations
}

func (a *APIcastOptionsProvider) getGatewayEmbeddedConfigSecret() (*v1.Secret, error) {
	if a.APIcastCR.Spec.EmbeddedConfigurationSecretRef == nil {
		return nil, nil
	}

	gatewayConfigSecretReference := a.APIcastCR.Spec.EmbeddedConfigurationSecretRef
	gatewayConfigSecretNamespace := a.APIcastCR.Namespace

	if gatewayConfigSecretReference.Name == "" {
		return nil, fmt.Errorf("Field 'Name' not specified for EmbeddedConfigurationSecretRef Secret Reference")
	}

	gatewayConfigSecretNamespacedName := types.NamespacedName{
		Name:      gatewayConfigSecretReference.Name,
		Namespace: gatewayConfigSecretNamespace,
	}

	gatewayConfigSecret := v1.Secret{}
	err := a.Client.Get(context.TODO(), gatewayConfigSecretNamespacedName, &gatewayConfigSecret)

	if err != nil {
		return nil, err
	}

	secretStringData := k8sutils.SecretStringDataFromData(gatewayConfigSecret)
	if _, ok := secretStringData[EmbeddedConfigurationSecretKey]; !ok {
		return nil, fmt.Errorf("Required key '%s' not found in secret '%s'", EmbeddedConfigurationSecretKey, gatewayConfigSecret.Name)
	}

	return &gatewayConfigSecret, err
}

func (a *APIcastOptionsProvider) getAdminPortalCredentialsSecret() (*v1.Secret, error) {
	if a.APIcastCR.Spec.AdminPortalCredentialsRef == nil {
		return nil, nil
	}

	adminPortalSecretReference := a.APIcastCR.Spec.AdminPortalCredentialsRef
	adminPortalNamespace := a.APIcastCR.Namespace

	if adminPortalSecretReference.Name == "" {
		return nil, fmt.Errorf("Field 'Name' not specified for AdminPortalCredentialsRef Secret Reference")
	}

	adminPortalCredentialsNamespacedName := types.NamespacedName{
		Name:      adminPortalSecretReference.Name,
		Namespace: adminPortalNamespace,
	}

	adminPortalCredentialsSecret := v1.Secret{}
	err := a.Client.Get(context.TODO(), adminPortalCredentialsNamespacedName, &adminPortalCredentialsSecret)

	if err != nil {
		return nil, err
	}

	secretStringData := k8sutils.SecretStringDataFromData(adminPortalCredentialsSecret)
	adminPortalURL, ok := secretStringData[AdminPortalURLAttributeName]
	if !ok {
		return nil, fmt.Errorf("Required key '%s' not found in secret '%s'", AdminPortalURLAttributeName, adminPortalCredentialsSecret.Name)
	}

	parsedURL, err := url.Parse(adminPortalURL)
	if err != nil {
		return nil, err
	}

	accessToken := parsedURL.User.Username()
	if accessToken == "" {
		return nil, fmt.Errorf("Access Token required in %s URL", AdminPortalURLAttributeName)
	}

	return &adminPortalCredentialsSecret, err
}

func (a *APIcastOptionsProvider) getHTTPSCertificateSecret() (*v1.Secret, error) {
	if a.APIcastCR.Spec.HTTPSCertificateSecretRef == nil {
		return nil, nil
	}

	errors := field.ErrorList{}
	specFldPath := field.NewPath("spec")
	httpsCertificateSecretRefFldPath := specFldPath.Child("httpsCertificateSecretRef")
	secretNameFldPath := httpsCertificateSecretRefFldPath.Child("name")

	ns := a.APIcastCR.Namespace

	if a.APIcastCR.Spec.HTTPSCertificateSecretRef.Name == "" {
		errors = append(errors, field.Required(secretNameFldPath, "secret name not provided"))
		return nil, errors.ToAggregate()
	}

	namespacedName := types.NamespacedName{
		Name:      a.APIcastCR.Spec.HTTPSCertificateSecretRef.Name,
		Namespace: ns,
	}

	secret := &v1.Secret{}
	err := a.Client.Get(context.TODO(), namespacedName, secret)

	if err != nil {
		// NotFoundError is also an error, it is required to exist
		return nil, err
	}

	if secret.Type != v1.SecretTypeTLS {
		errors = append(errors, field.Invalid(httpsCertificateSecretRefFldPath, a.APIcastCR.Spec.HTTPSCertificateSecretRef, "Required kubernetes.io/tls secret type"))
		return nil, errors.ToAggregate()
	}

	if _, ok := secret.Data[v1.TLSCertKey]; !ok {
		errors = append(errors, field.Required(httpsCertificateSecretRefFldPath, fmt.Sprintf("Required secret key, %s not found", v1.TLSCertKey)))
		return nil, errors.ToAggregate()
	}

	if _, ok := secret.Data[v1.TLSPrivateKeyKey]; !ok {
		errors = append(errors, field.Required(httpsCertificateSecretRefFldPath, fmt.Sprintf("Required secret key, %s not found", v1.TLSPrivateKeyKey)))
		return nil, errors.ToAggregate()
	}

	return secret, err
}

func (a *APIcastOptionsProvider) validateCustomPolicySecret(nn types.NamespacedName) error {
	secret := &v1.Secret{}
	err := a.Client.Get(context.TODO(), nn, secret)

	if err != nil {
		// NotFoundError is also an error, it is required to exist
		return err
	}

	if _, ok := secret.Data["init.lua"]; !ok {
		return fmt.Errorf("Required secret key, %s not found", "init.lua")
	}

	if _, ok := secret.Data["apicast-policy.json"]; !ok {
		return fmt.Errorf("Required secret key, %s not found", "apicast-policy.json")
	}

	return nil
}
