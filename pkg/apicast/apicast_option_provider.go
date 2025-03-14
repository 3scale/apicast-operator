package apicast

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"path"
	"sort"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/client"

	appsv1alpha1 "github.com/3scale/apicast-operator/apis/apps/v1alpha1"
	"github.com/3scale/apicast-operator/pkg/helper"
	"github.com/3scale/apicast-operator/pkg/k8sutils"
)

const (
	AdmPortalSecretResverAnnotation            = "apicast.apps.3scale.net/admin-portal-secret-resource-version"
	GatewayConfigurationSecretResverAnnotation = "apicast.apps.3scale.net/gateway-configuration-secret-resource-version"
	HttpsCertSecretResverAnnotation            = "apicast.apps.3scale.net/https-cert-secret-resource-version"
	OpenTracingSecretResverAnnotation          = "apicast.apps.3scale.net/opentracing-secret-resource-version"
	OpenTelemetrySecretResverAnnotation        = "apicast.apps.3scale.net/opentelemetry-secret-resource-version"
	CustomEnvSecretResverAnnotationPrefix      = "apicast.apps.3scale.net/customenv-secret-resource-version-"
	CustomPoliciesSecretResverAnnotationPrefix = "apicast.apps.3scale.net/custompolicy-secret-resource-version-"
	APPLABEL                                   = "apicast"
)

type APIcastOptionsProvider struct {
	APIcastCR      *appsv1alpha1.APIcast
	APIcastOptions *APIcastOptions
	Client         client.Client
}

func APIcastDeploymentName(cr *appsv1alpha1.APIcast) string {
	if cr == nil {
		return ""
	}

	return fmt.Sprintf("apicast-%s", cr.Name)
}

func NewApicastOptionsProvider(cr *appsv1alpha1.APIcast, cl client.Client) *APIcastOptionsProvider {
	return &APIcastOptionsProvider{
		APIcastCR:      cr,
		APIcastOptions: NewAPIcastOptions(),
		Client:         cl,
	}
}

func (a *APIcastOptionsProvider) GetApicastOptions(ctx context.Context) (*APIcastOptions, error) {
	a.APIcastOptions.Namespace = a.APIcastCR.Namespace
	a.APIcastOptions.Owner = a.APIcastCR.GetOwnerReference()

	a.APIcastOptions.DeploymentName = APIcastDeploymentName(a.APIcastCR)
	a.APIcastOptions.ServiceName = APIcastDeploymentName(a.APIcastCR)

	a.APIcastOptions.Replicas = 1
	if a.APIcastCR.Spec.Replicas != nil {
		a.APIcastOptions.Replicas = int32(*a.APIcastCR.Spec.Replicas)
	}

	a.APIcastOptions.Hpa = a.APIcastCR.Spec.Hpa

	a.APIcastOptions.ServiceAccountName = "default"
	if a.APIcastCR.Spec.ServiceAccount != nil {
		a.APIcastOptions.ServiceAccountName = *a.APIcastCR.Spec.ServiceAccount
	}

	a.APIcastOptions.Image = GetDefaultImageVersion()
	if a.APIcastCR.Spec.Image != nil {
		a.APIcastOptions.Image = *a.APIcastCR.Spec.Image
	}

	a.APIcastOptions.PodLabelSelector = a.podLabelSelector(a.APIcastOptions.DeploymentName)
	a.APIcastOptions.CommonLabels = a.commonLabels()
	a.APIcastOptions.PodTemplateLabels = a.podTemplateLabels(a.APIcastOptions.PodLabelSelector)

	a.APIcastOptions.ExposedHost = ExposedHost{}
	if a.APIcastCR.Spec.ExposedHost != nil {
		a.APIcastOptions.ExposedHost.Host = a.APIcastCR.Spec.ExposedHost.Host
		a.APIcastOptions.ExposedHost.TLS = a.APIcastCR.Spec.ExposedHost.TLS
	}

	adminPortalCredentialsSecret, err := a.getAdminPortalCredentialsSecret(ctx)
	if err != nil {
		return nil, err
	}
	a.APIcastOptions.AdminPortalCredentialsSecret = adminPortalCredentialsSecret

	gatewayConfigurationSecret, err := a.getGatewayEmbeddedConfigSecret(ctx)
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
	a.APIcastOptions.ServiceCacheSize = a.APIcastCR.Spec.ServiceCacheSize
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
	httpsCertificateSecret, err := a.getHTTPSCertificateSecret(ctx)
	if err != nil {
		return nil, err
	}
	a.APIcastOptions.HTTPSCertificateSecret = httpsCertificateSecret

	caCertificateSecret, err := a.getCACertificateSecret(ctx)
	if err != nil {
		return nil, err
	}
	a.APIcastOptions.CACertificateSecret = caCertificateSecret

	// Resource requirements
	resourceRequirements := DefaultResourceRequirements(a.APIcastCR.Spec.Hpa)

	// Apply Resources configuration from APICast CR if available
	if a.APIcastCR.Spec.Resources != nil {
		resourceRequirements = *a.APIcastCR.Spec.Resources
	}
	a.APIcastOptions.ResourceRequirements = resourceRequirements

	a.APIcastOptions.Workers = a.APIcastCR.Spec.Workers
	a.APIcastOptions.Timezone = a.APIcastCR.Spec.Timezone

	a.APIcastOptions.AllProxy = a.APIcastCR.Spec.AllProxy
	a.APIcastOptions.HTTPProxy = a.APIcastCR.Spec.HTTPProxy
	a.APIcastOptions.HTTPSProxy = a.APIcastCR.Spec.HTTPSProxy
	a.APIcastOptions.NoProxy = a.APIcastCR.Spec.NoProxy

	for idx, customPolicySpec := range a.APIcastCR.Spec.CustomPolicies {
		namespacedName := types.NamespacedName{
			Name:      customPolicySpec.SecretRef.Name, // CR Validation ensures not nil
			Namespace: a.APIcastCR.Namespace,
		}
		secret, err := a.validateCustomPolicySecret(ctx, namespacedName)
		if err != nil {
			errors := field.ErrorList{}
			customPoliciesIdxFldPath := field.NewPath("spec").Child("customPolicies").Index(idx)
			errors = append(errors, field.Invalid(customPoliciesIdxFldPath, customPolicySpec, err.Error()))
			return nil, errors.ToAggregate()
		}

		a.APIcastOptions.CustomPolicies = append(a.APIcastOptions.CustomPolicies, CustomPolicy{
			Name:    customPolicySpec.Name,
			Version: customPolicySpec.Version,
			Secret:  secret,
		})
	}

	a.APIcastOptions.ExtendedMetrics = a.APIcastCR.Spec.ExtendedMetrics

	for idx, customEnvSpec := range a.APIcastCR.Spec.CustomEnvironments {
		namespacedName := types.NamespacedName{
			Name:      customEnvSpec.SecretRef.Name, // CR Validation ensures not nil
			Namespace: a.APIcastCR.Namespace,
		}

		secret, err := a.customEnvSecret(ctx, namespacedName)
		if err != nil {
			errors := field.ErrorList{}
			customEnvsIdxFldPath := field.NewPath("spec").Child("customEnvironments").Index(idx)
			errors = append(errors, field.Invalid(customEnvsIdxFldPath, customEnvSpec, err.Error()))
			return nil, errors.ToAggregate()
		}

		a.APIcastOptions.CustomEnvironments = append(a.APIcastOptions.CustomEnvironments, secret)
	}

	tracingOptions, err := a.getTracingConfigOptions(ctx)
	if err != nil {
		return nil, err
	}
	a.APIcastOptions.TracingConfig = tracingOptions

	//
	otelConfig, err := a.getOpenTelemetryConfig(ctx)
	if err != nil {
		return nil, err
	}
	a.APIcastOptions.Opentelemetry = otelConfig

	return a.APIcastOptions, a.APIcastOptions.Validate()
}

func (a *APIcastOptionsProvider) getTracingConfigOptions(ctx context.Context) (TracingConfig, error) {
	tracingIsEnabled := a.APIcastCR.OpenTracingIsEnabled()
	res := TracingConfig{
		Enabled:        tracingIsEnabled,
		TracingLibrary: DefaultTracingLibrary,
	}
	if tracingIsEnabled {
		openTracingConfigSpec := a.APIcastCR.Spec.OpenTracing
		if openTracingConfigSpec.TracingLibrary != nil {
			// For now only "jaeger" is accepted" as the tracing library
			if *openTracingConfigSpec.TracingLibrary != DefaultTracingLibrary {
				tracingLibraryFldPath := field.NewPath("spec").Child("openTracing").Child("tracingLibrary")
				errors := field.ErrorList{}
				errors = append(errors, field.Invalid(tracingLibraryFldPath, openTracingConfigSpec, "invalid tracing library specified"))
				return res, errors.ToAggregate()
			}
			res.TracingLibrary = *a.APIcastCR.Spec.OpenTracing.TracingLibrary
		}
		if openTracingConfigSpec.TracingConfigSecretRef != nil {
			namespacedName := types.NamespacedName{
				Name:      openTracingConfigSpec.TracingConfigSecretRef.Name, // CR Validation ensures not nil
				Namespace: a.APIcastCR.Namespace,
			}
			secret, err := a.validateTracingConfigSecret(ctx, namespacedName)
			if err != nil {
				errors := field.ErrorList{}
				tracingConfigFldPath := field.NewPath("spec").Child("openTracing").Child("tracingConfigSecretRef")
				errors = append(errors, field.Invalid(tracingConfigFldPath, openTracingConfigSpec, err.Error()))
				return res, errors.ToAggregate()
			}
			res.Secret = secret
		}
	}

	return res, nil
}

func (a *APIcastOptionsProvider) getGatewayEmbeddedConfigSecret(ctx context.Context) (*v1.Secret, error) {
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
	err := a.Client.Get(ctx, gatewayConfigSecretNamespacedName, &gatewayConfigSecret)

	if err != nil {
		return nil, err
	}

	secretStringData := k8sutils.SecretStringDataFromData(&gatewayConfigSecret)
	if _, ok := secretStringData[EmbeddedConfigurationSecretKey]; !ok {
		return nil, fmt.Errorf("Required key '%s' not found in secret '%s'", EmbeddedConfigurationSecretKey, gatewayConfigSecret.Name)
	}

	return &gatewayConfigSecret, err
}

func (a *APIcastOptionsProvider) getAdminPortalCredentialsSecret(ctx context.Context) (*v1.Secret, error) {
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
	err := a.Client.Get(ctx, adminPortalCredentialsNamespacedName, &adminPortalCredentialsSecret)

	if err != nil {
		return nil, err
	}

	secretStringData := k8sutils.SecretStringDataFromData(&adminPortalCredentialsSecret)
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

func (a *APIcastOptionsProvider) getHTTPSCertificateSecret(ctx context.Context) (*v1.Secret, error) {
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
	err := a.Client.Get(ctx, namespacedName, secret)

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

func (a *APIcastOptionsProvider) getCACertificateSecret(ctx context.Context) (*v1.Secret, error) {
	if a.APIcastCR.Spec.CACertificateSecretRef == nil {
		return nil, nil
	}

	errors := field.ErrorList{}
	specFldPath := field.NewPath("spec")
	caCertificateSecretRefFldPath := specFldPath.Child("caCertificateSecretRef")
	secretNameFldPath := caCertificateSecretRefFldPath.Child("name")

	ns := a.APIcastCR.Namespace

	if a.APIcastCR.Spec.CACertificateSecretRef.Name == "" {
		errors = append(errors, field.Required(secretNameFldPath, "secret name not provided"))
		return nil, errors.ToAggregate()
	}

	namespacedName := types.NamespacedName{
		Name:      a.APIcastCR.Spec.CACertificateSecretRef.Name,
		Namespace: ns,
	}

	secret := &v1.Secret{}
	err := a.Client.Get(ctx, namespacedName, secret)

	if err != nil {
		// NotFoundError is also an error, it is required to exist
		return nil, err
	}

	return secret, nil
}

func (a *APIcastOptionsProvider) validateCustomPolicySecret(ctx context.Context, nn types.NamespacedName) (*v1.Secret, error) {
	secret := &v1.Secret{}
	err := a.Client.Get(ctx, nn, secret)

	if err != nil {
		// NotFoundError is also an error, it is required to exist
		return nil, err
	}

	if _, ok := secret.Data["init.lua"]; !ok {
		return nil, fmt.Errorf("Required secret key, %s not found", "init.lua")
	}

	if _, ok := secret.Data["apicast-policy.json"]; !ok {
		return nil, fmt.Errorf("Required secret key, %s not found", "apicast-policy.json")
	}

	return secret, nil
}

func (a *APIcastOptionsProvider) customEnvSecret(ctx context.Context, nn types.NamespacedName) (*v1.Secret, error) {
	secret := &v1.Secret{}
	err := a.Client.Get(ctx, nn, secret)

	if err != nil {
		// NotFoundError is also an error, it is required to exist
		return nil, err
	}

	if len(secret.Data) == 0 {
		return nil, errors.New("empty secret")
	}

	return secret, nil
}

func (a *APIcastOptionsProvider) validateTracingConfigSecret(ctx context.Context, nn types.NamespacedName) (*v1.Secret, error) {
	secret := &v1.Secret{}
	err := a.Client.Get(ctx, nn, secret)

	if err != nil {
		// NotFoundError is also an error, it is required to exist
		return nil, err
	}

	if _, ok := secret.Data[TracingConfigSecretKey]; !ok {
		return nil, fmt.Errorf("Required secret key, %s not found", TracingConfigSecretKey)
	}

	return secret, nil
}

func (a *APIcastOptionsProvider) commonLabels() map[string]string {
	return map[string]string{
		"app":                  APPLABEL,
		"threescale_component": "apicast",
	}
}

func (a *APIcastOptionsProvider) podTemplateLabels(labelSelector map[string]string) map[string]string {
	meteringLabels := helper.MeteringLabels(helper.ApplicationType)

	// merge maps

	result := make(map[string]string)

	for k, v := range meteringLabels {
		result[k] = v
	}

	for k, v := range labelSelector {
		result[k] = v
	}

	return result
}

func (a *APIcastOptionsProvider) podLabelSelector(deploymentName string) map[string]string {
	return map[string]string{
		"deployment": deploymentName,
	}
}

func (a *APIcastOptionsProvider) getOpenTelemetryConfig(ctx context.Context) (OpentelemetryConfig, error) {
	res := OpentelemetryConfig{
		Enabled: a.APIcastCR.OpenTelemetryEnabled(),
	}

	if !res.Enabled {
		return res, nil
	}

	// In the APIcast CR validation step it is checked that when enabled, the secret ref is not nil
	// Adding this to avoid panics
	if a.APIcastCR.Spec.OpenTelemetry.TracingConfigSecretRef == nil || a.APIcastCR.Spec.OpenTelemetry.TracingConfigSecretRef.Name == "" {
		fldPath := field.NewPath("spec").Child("openTelemetry").Child("tracingConfigSecretRef")
		errors := append(field.ErrorList{}, field.Invalid(fldPath, a.APIcastCR.Spec.OpenTelemetry.TracingConfigSecretRef, "tracing config secret name is empty"))
		return res, errors.ToAggregate()
	}

	res.SecretName = a.APIcastCR.Spec.OpenTelemetry.TracingConfigSecretRef.Name

	if a.APIcastCR.Spec.OpenTelemetry.TracingConfigSecretKey != nil &&
		*a.APIcastCR.Spec.OpenTelemetry.TracingConfigSecretKey != "" {
		res.ConfigFile = path.Join(OpentelemetryConfigMountBasePath, *a.APIcastCR.Spec.OpenTelemetry.TracingConfigSecretKey)
		return res, nil
	}

	// Read secret and get first key in lexicographical order.
	// Defining some order is required because maps do not provide order semantics and
	// key consistency is required accross reconcile loops
	otelSecretKey := client.ObjectKey{
		Name:      a.APIcastCR.Spec.OpenTelemetry.TracingConfigSecretRef.Name,
		Namespace: a.APIcastCR.Namespace,
	}

	secret := &v1.Secret{}
	err := a.Client.Get(ctx, otelSecretKey, secret)
	if err != nil {
		// NotFoundError is also an error, it is required to exist
		fldPath := field.NewPath("spec").Child("openTelemetry").Child("tracingConfigSecretRef")
		errors := append(field.ErrorList{}, field.Invalid(fldPath, a.APIcastCR.Spec.OpenTelemetry.TracingConfigSecretRef, err.Error()))
		return res, errors.ToAggregate()
	}

	secretKeys := helper.MapKeys(k8sutils.SecretStringDataFromData(secret))
	if len(secretKeys) == 0 {
		fldPath := field.NewPath("spec").Child("openTelemetry").Child("tracingConfigSecretRef")
		errors := append(field.ErrorList{}, field.Invalid(fldPath, a.APIcastCR.Spec.OpenTelemetry.TracingConfigSecretRef, "secret is empty, no key found"))
		return res, errors.ToAggregate()
	}

	sort.Strings(secretKeys)

	res.ConfigFile = path.Join(OpentelemetryConfigMountBasePath, secretKeys[0])

	return res, nil
}
