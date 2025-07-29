package apicast

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path"
	"reflect"
	"sort"
	"strconv"
	"strings"

	appsv1alpha1 "github.com/3scale/apicast-operator/apis/apps/v1alpha1"
	"github.com/3scale/apicast-operator/pkg/helper"
	"github.com/3scale/apicast-operator/pkg/k8sutils"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	AdminPortalURLAttributeName       = "AdminPortalURL"
	DefaultManagementPort       int32 = 8090
	DefaultMetricsPort          int32 = 9421
	DefaultTracingLibrary             = "jaeger"
	TracingConfigSecretKey            = "config"
)

const (
	EmbeddedConfigurationMountPath  = "/tmp/gateway-configuration-volume"
	EmbeddedConfigurationVolumeName = "gateway-configuration-volume"
	EmbeddedConfigurationSecretKey  = "config.json"
)

const (
	HTTPSCertificatesMountPath  = "/var/run/secrets/apicast"
	HTTPSCertificatesVolumeName = "https-certificates"
	CACertificatesSecretKey     = "ca-bundle.crt"
	CACertificatesVolumeName    = "ca-certificate"
	CustomPoliciesMountBasePath = "/opt/app-root/src/policies"
	CustomEnvsMountBasePath     = "/opt/app-root/src/custom-environments"
	TracingConfigMountBasePath  = "/opt/app-root/src/tracing-configs"
)

const (
	OpentelemetryConfigurationVolumeName = "otel-volume"
	OpentelemetryConfigMountBasePath     = "/opt/app-root/src/otel-configs"
)

const (
	HashedSecretName = "hashed-secret-data"
)

type APIcast struct {
	options *APIcastOptions
}

func NewAPIcast(opts *APIcastOptions) *APIcast {
	return &APIcast{options: opts}
}

func Factory(ctx context.Context, cr *appsv1alpha1.APIcast, cl client.Client) (*APIcast, error) {
	optsProvider := NewApicastOptionsProvider(cr, cl)
	opts, err := optsProvider.GetApicastOptions(ctx)
	if err != nil {
		return nil, err
	}
	return NewAPIcast(opts), nil
}

func (a *APIcast) deploymentVolumeMounts() []v1.VolumeMount {
	var volumeMounts []v1.VolumeMount
	if a.options.GatewayConfigurationSecret != nil {
		volumeMounts = append(volumeMounts, v1.VolumeMount{
			Name:      EmbeddedConfigurationVolumeName,
			MountPath: EmbeddedConfigurationMountPath,
			ReadOnly:  true,
		})
	}

	if a.options.HTTPSCertificateSecret != nil {
		volumeMounts = append(volumeMounts, v1.VolumeMount{
			Name:      HTTPSCertificatesVolumeName,
			MountPath: HTTPSCertificatesMountPath,
			ReadOnly:  true,
		})
	}

	// Use the same mount path with https certificate
	if a.options.CACertificateSecret != nil {
		volumeMounts = append(volumeMounts, v1.VolumeMount{
			Name:      CACertificatesVolumeName,
			MountPath: HTTPSCertificatesMountPath,
			ReadOnly:  true,
		})
	}

	for _, customPolicy := range a.options.CustomPolicies {
		volumeMounts = append(volumeMounts, v1.VolumeMount{
			Name:      policyVolumeName(customPolicy),
			MountPath: path.Join(CustomPoliciesMountBasePath, customPolicy.Name, customPolicy.Version),
			ReadOnly:  true,
		})
	}

	for _, customEnvSecret := range a.options.CustomEnvironments {
		volumeMounts = append(volumeMounts, v1.VolumeMount{
			Name:      customEnvVolumeName(customEnvSecret),
			MountPath: path.Join(CustomEnvsMountBasePath, customEnvSecret.GetName()),
			ReadOnly:  true,
		})
	}

	if a.options.TracingConfig.Enabled && a.options.TracingConfig.Secret != nil {
		volumeMounts = append(volumeMounts, v1.VolumeMount{
			Name:      tracingConfigVolumeName(a.options.TracingConfig.TracingLibrary, a.options.TracingConfig.Secret.Name),
			MountPath: TracingConfigMountBasePath,
		})
	}

	if a.options.Opentelemetry.Enabled {
		volumeMounts = append(volumeMounts, v1.VolumeMount{
			Name:      OpentelemetryConfigurationVolumeName,
			MountPath: OpentelemetryConfigMountBasePath,
			ReadOnly:  true,
		})
	}

	return volumeMounts
}

func policyVolumeName(cp CustomPolicy) string {
	return fmt.Sprintf("policy-%s-%s", helper.DNS1123Name(cp.Version), helper.DNS1123Name(cp.Name))
}

func customEnvVolumeName(secret *v1.Secret) string {
	return fmt.Sprintf("custom-env-%s", secret.GetName())
}

func tracingConfigVolumeName(tracingLibrary, secretName string) string {
	return fmt.Sprintf("tracing-config-%s-%s", tracingLibrary, secretName)
}

func (a *APIcast) deploymentVolumes() []v1.Volume {
	var volumes []v1.Volume
	if a.options.GatewayConfigurationSecret != nil {
		volumes = append(volumes, v1.Volume{
			Name: EmbeddedConfigurationVolumeName,
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: a.options.GatewayConfigurationSecret.Name,
					Items: []v1.KeyToPath{
						{
							Key:  EmbeddedConfigurationSecretKey,
							Path: EmbeddedConfigurationSecretKey,
						},
					},
				},
			},
		})
	}

	if a.options.HTTPSCertificateSecret != nil {
		volumes = append(volumes, v1.Volume{
			Name: HTTPSCertificatesVolumeName,
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: a.options.HTTPSCertificateSecret.Name,
				},
			},
		})
	}

	if a.options.CACertificateSecret != nil {
		volumes = append(volumes, v1.Volume{
			Name: CACertificatesVolumeName,
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: a.options.CACertificateSecret.Name,
					Items: []v1.KeyToPath{
						{
							Key:  CACertificatesSecretKey,
							Path: "ca-bundle.crt", // Map the secret key to the ca-bundle.crt file in the container
						},
					},
				},
			},
		})
	}

	for _, customPolicy := range a.options.CustomPolicies {
		volumes = append(volumes, v1.Volume{
			Name: policyVolumeName(customPolicy),
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: customPolicy.Secret.Name,
				},
			},
		})
	}

	for _, customEnvSecret := range a.options.CustomEnvironments {
		volumes = append(volumes, v1.Volume{
			Name: customEnvVolumeName(customEnvSecret),
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: customEnvSecret.GetName(),
				},
			},
		})
	}

	if a.options.TracingConfig.Enabled && a.options.TracingConfig.Secret != nil {
		volumes = append(volumes, v1.Volume{
			Name: tracingConfigVolumeName(a.options.TracingConfig.TracingLibrary, a.options.TracingConfig.Secret.Name),
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: a.options.TracingConfig.Secret.Name,
					Items: []v1.KeyToPath{
						{
							Key:  TracingConfigSecretKey,
							Path: tracingConfigVolumeName(a.options.TracingConfig.TracingLibrary, a.options.TracingConfig.Secret.Name),
						},
					},
				},
			},
		})
	}

	if a.options.Opentelemetry.Enabled {
		volumes = append(volumes, v1.Volume{
			Name: OpentelemetryConfigurationVolumeName,
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: a.options.Opentelemetry.SecretName,
					Optional:   &[]bool{false}[0],
				},
			},
		})
	}

	return volumes
}

func (a *APIcast) deploymentEnv() []v1.EnvVar {
	var env []v1.EnvVar

	if a.options.AdminPortalCredentialsSecret != nil {
		env = append(env, k8sutils.EnvVarFromSecretKey("THREESCALE_PORTAL_ENDPOINT", a.options.AdminPortalCredentialsSecret.Name, AdminPortalURLAttributeName))
	}

	if a.options.DeploymentEnvironment != nil {
		env = append(env, k8sutils.EnvVarFromValue("THREESCALE_DEPLOYMENT_ENV", *a.options.DeploymentEnvironment))
	}

	if a.options.DNSResolverAddress != nil {
		env = append(env, k8sutils.EnvVarFromValue("RESOLVER", *a.options.DNSResolverAddress))
	}

	if a.options.EnabledServices != nil {
		joinedStr := strings.Join(a.options.EnabledServices, ",")
		if joinedStr != "" {
			env = append(env, k8sutils.EnvVarFromValue("APICAST_SERVICES_LIST", joinedStr))
		}
	}

	if a.options.ConfigurationLoadMode != nil {
		env = append(env, k8sutils.EnvVarFromValue("APICAST_CONFIGURATION_LOADER", *a.options.ConfigurationLoadMode))
	}

	if a.options.LogLevel != nil {
		env = append(env, k8sutils.EnvVarFromValue("APICAST_LOG_LEVEL", *a.options.LogLevel))
	}

	if a.options.PathRoutingEnabled != nil {
		env = append(env, k8sutils.EnvVarFromValue("APICAST_PATH_ROUTING", strconv.FormatBool(*a.options.PathRoutingEnabled)))
	}

	if a.options.ResponseCodesIncluded != nil {
		env = append(env, k8sutils.EnvVarFromValue("APICAST_RESPONSE_CODES", strconv.FormatBool(*a.options.ResponseCodesIncluded)))
	}

	if a.options.CacheConfigurationSeconds != nil {
		env = append(env, k8sutils.EnvVarFromValue("APICAST_CONFIGURATION_CACHE", strconv.FormatInt(*a.options.CacheConfigurationSeconds, 10)))
	}

	if a.options.ManagementAPIScope != nil {
		env = append(env, k8sutils.EnvVarFromValue("APICAST_MANAGEMENT_API", *a.options.ManagementAPIScope))
	}

	if a.options.OpenSSLPeerVerificationEnabled != nil {
		env = append(env, k8sutils.EnvVarFromValue("OPENSSL_VERIFY", strconv.FormatBool(*a.options.OpenSSLPeerVerificationEnabled)))
	}

	if a.options.GatewayConfigurationSecret != nil {
		env = append(env, v1.EnvVar{
			Name:  "THREESCALE_CONFIG_FILE",
			Value: EmbeddedConfigurationMountPath + "/" + EmbeddedConfigurationSecretKey,
		})
	}

	if a.options.UpstreamRetryCases != nil {
		env = append(env, k8sutils.EnvVarFromValue("APICAST_UPSTREAM_RETRY_CASES", *a.options.UpstreamRetryCases))
	}

	if a.options.CacheMaxTime != nil {
		env = append(env, k8sutils.EnvVarFromValue("APICAST_CACHE_MAX_TIME", *a.options.CacheMaxTime))
	}

	if a.options.CacheStatusCodes != nil {
		env = append(env, k8sutils.EnvVarFromValue("APICAST_CACHE_STATUS_CODES", *a.options.CacheStatusCodes))
	}

	if a.options.ServiceCacheSize != nil {
		env = append(env, k8sutils.EnvVarFromValue("APICAST_SERVICE_CACHE_SIZE", fmt.Sprintf("%d", *a.options.ServiceCacheSize)))
	}

	if a.options.OidcLogLevel != nil {
		env = append(env, k8sutils.EnvVarFromValue("APICAST_OIDC_LOG_LEVEL", *a.options.OidcLogLevel))
	}

	if a.options.LoadServicesWhenNeeded != nil {
		env = append(env, k8sutils.EnvVarFromValue("APICAST_LOAD_SERVICES_WHEN_NEEDED", strconv.FormatBool(*a.options.LoadServicesWhenNeeded)))
	}

	if a.options.ServicesFilterByURL != nil {
		env = append(env, k8sutils.EnvVarFromValue("APICAST_SERVICES_FILTER_BY_URL", *a.options.ServicesFilterByURL))
	}

	for serviceID, serviceVersion := range a.options.ServiceConfigurationVersionOverride {
		env = append(env, k8sutils.EnvVarFromValue(fmt.Sprintf("APICAST_SERVICE_%s_CONFIGURATION_VERSION", serviceID), serviceVersion))
	}

	if a.options.HTTPSPort != nil {
		env = append(env, k8sutils.EnvVarFromValue("APICAST_HTTPS_PORT", strconv.FormatInt(int64(*a.options.HTTPSPort), 10)))
	}

	if a.options.HTTPSVerifyDepth != nil {
		env = append(env, k8sutils.EnvVarFromValue("APICAST_HTTPS_VERIFY_DEPTH", strconv.FormatInt(*a.options.HTTPSVerifyDepth, 10)))
	}

	if a.options.HTTPSCertificateSecret != nil {
		env = append(env,
			k8sutils.EnvVarFromValue("APICAST_HTTPS_CERTIFICATE", fmt.Sprintf("%s/%s", HTTPSCertificatesMountPath, v1.TLSCertKey)),
			k8sutils.EnvVarFromValue("APICAST_HTTPS_CERTIFICATE_KEY", fmt.Sprintf("%s/%s", HTTPSCertificatesMountPath, v1.TLSPrivateKeyKey)))
	}

	if a.options.CACertificateSecret != nil {
		env = append(env,
			k8sutils.EnvVarFromValue("SSL_CERT_FILE", path.Join(HTTPSCertificatesMountPath, "ca-bundle.crt")))
	}

	if a.options.Workers != nil {
		env = append(env, k8sutils.EnvVarFromValue("APICAST_WORKERS", strconv.Itoa(int(*a.options.Workers))))
	}

	if a.options.Timezone != nil {
		env = append(env, k8sutils.EnvVarFromValue("TZ", *a.options.Timezone))
	}

	if a.options.ExtendedMetrics != nil {
		env = append(env, k8sutils.EnvVarFromValue("APICAST_EXTENDED_METRICS", strconv.FormatBool(*a.options.ExtendedMetrics)))
	}

	var customEnvPaths []string
	for _, customEnvSecret := range a.options.CustomEnvironments {
		for fileKey := range customEnvSecret.Data {
			customEnvPaths = append(customEnvPaths, path.Join(CustomEnvsMountBasePath, customEnvSecret.GetName(), fileKey))
		}
	}

	if len(customEnvPaths) > 0 {
		env = append(env, k8sutils.EnvVarFromValue("APICAST_ENVIRONMENT", strings.Join(customEnvPaths, ":")))
	}

	if a.options.TracingConfig.Enabled {
		env = append(env, k8sutils.EnvVarFromValue("OPENTRACING_TRACER", a.options.TracingConfig.TracingLibrary))

		if a.options.TracingConfig.Secret != nil {
			env = append(env, k8sutils.EnvVarFromValue("OPENTRACING_CONFIG", strings.Join([]string{TracingConfigMountBasePath, tracingConfigVolumeName(a.options.TracingConfig.TracingLibrary, a.options.TracingConfig.Secret.Name)}, "/")))
		}
	}

	if a.options.AllProxy != nil {
		env = append(env, k8sutils.EnvVarFromValue("ALL_PROXY", *a.options.AllProxy))
	}

	if a.options.HTTPProxy != nil {
		env = append(env, k8sutils.EnvVarFromValue("HTTP_PROXY", *a.options.HTTPProxy))
	}

	if a.options.HTTPSProxy != nil {
		env = append(env, k8sutils.EnvVarFromValue("HTTPS_PROXY", *a.options.HTTPSProxy))
	}

	if a.options.NoProxy != nil {
		env = append(env, k8sutils.EnvVarFromValue("NO_PROXY", *a.options.NoProxy))
	}

	if a.options.Opentelemetry.Enabled {
		env = append(env, k8sutils.EnvVarFromValue("OPENTELEMETRY", "1"))
		env = append(env, k8sutils.EnvVarFromValue("OPENTELEMETRY_CONFIG", a.options.Opentelemetry.ConfigFile))
	}

	return env
}

func (a *APIcast) Deployment(ctx context.Context, k8sclient client.Client) (*appsv1.Deployment, error) {
	watchedSecretAnnotations, err := a.computeWatchedSecretAnnotations(ctx, k8sclient)
	if err != nil {
		return nil, err
	}

	deployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      a.options.DeploymentName,
			Namespace: a.options.Namespace,
			Labels:    a.options.CommonLabels,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: a.options.PodLabelSelector,
			},
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RollingUpdateDeploymentStrategyType,
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      a.options.PodTemplateLabels,
					Annotations: a.podAnnotations(watchedSecretAnnotations),
				},
				Spec: v1.PodSpec{
					ServiceAccountName: a.options.ServiceAccountName,
					Volumes:            a.deploymentVolumes(),
					Containers: []v1.Container{
						{
							Name:            a.options.DeploymentName,
							Ports:           a.containerPorts(),
							Image:           a.options.Image,
							ImagePullPolicy: v1.PullAlways, // This is different than the currently used which is IfNotPresent
							Resources:       a.options.ResourceRequirements,
							LivenessProbe:   a.livenessProbe(),
							ReadinessProbe:  a.readinessProbe(),
							VolumeMounts:    a.deploymentVolumeMounts(),
							// Env takes precedence with respect to EnvFrom on duplicated
							// var values
							Env: a.deploymentEnv(),
						},
					},
				},
			},
		},
	}

	// Only configure deployment replicas when HPA is disabled, otherwise, HPA is in charge of replicas count
	if !a.options.Hpa {
		deployment.Spec.Replicas = &a.options.Replicas
	}

	addOwnerRefToObject(deployment, *a.options.Owner)
	return deployment, nil
}

func (a *APIcast) computeWatchedSecretAnnotations(ctx context.Context, k8sclient client.Client) (map[string]string, error) {
	// First get the initial annotations
	uncheckedAnnotations, err := a.getWatchedSecretAnnotations(ctx, k8sclient)
	if err != nil {
		return nil, err
	}

	// Then get the deployment (if it exists) to compare the existing annotations
	deployment := &appsv1.Deployment{}
	deploymentKey := client.ObjectKey{
		Name:      a.options.DeploymentName,
		Namespace: a.options.Namespace,
	}
	err = k8sclient.Get(ctx, deploymentKey, deployment)
	if err != nil && !apierrors.IsNotFound(err) {
		return nil, err
	}
	// If the deployment doesn't exist yet then just return the uncheckedAnnotations because there's nothing to compare to
	if apierrors.IsNotFound(err) {
		return uncheckedAnnotations, nil
	}

	// Next get the master hashed secret (if it exists) to compare the secret hashes
	hashedSecret := &v1.Secret{}
	hashedSecretKey := client.ObjectKey{
		Name:      HashedSecretName,
		Namespace: a.options.Namespace,
	}
	err = k8sclient.Get(ctx, hashedSecretKey, hashedSecret)
	if err != nil && !apierrors.IsNotFound(err) {
		return nil, err
	}
	// If the master hashed secret doesn't exist yet then just return the uncheckedAnnotations because there's nothing to compare to
	if apierrors.IsNotFound(err) {
		return uncheckedAnnotations, nil
	}

	existingPodAnnotations := deployment.Spec.Template.Annotations

	// First check if the annotations match, if they do then there's no need to check the hash
	if !reflect.DeepEqual(existingPodAnnotations, uncheckedAnnotations) {
		reconciledAnnotations := map[string]string{}

		// Loop through the annotations to see if the secret data has actually changed
		for key, resourceVersion := range uncheckedAnnotations {
			if existingPodAnnotations[key] == "" {
				reconciledAnnotations[key] = resourceVersion // If this is the first time adding the annotation then use the new resourceVersion
			} else if existingPodAnnotations[key] != resourceVersion && a.hasSecretHashChanged(ctx, k8sclient, key, hashedSecret) {
				reconciledAnnotations[key] = resourceVersion // Else if the resourceVersions don't match and the hash has changed then use the new resourceVersion
			} else {
				reconciledAnnotations[key] = existingPodAnnotations[key] // Otherwise keep the existing resourceVersion
			}
		}

		return reconciledAnnotations, nil
	}
	return uncheckedAnnotations, nil // No difference with existing annotations so can return uncheckedAnnotations
}

func (a *APIcast) getWatchedSecretAnnotations(ctx context.Context, k8sclient client.Client) (map[string]string, error) {
	// admin portal secret
	// gateway conf secret
	// https certificate secret
	// tracing config secret
	// telemetry config secret
	// custom policy secret(s)
	// custom env secret(s)

	annotations := map[string]string{}

	if a.options.AdminPortalCredentialsSecret != nil && k8sutils.IsSecretWatchedByApicast(a.options.AdminPortalCredentialsSecret) {
		annotations[AdmPortalSecretResverAnnotation] = a.options.AdminPortalCredentialsSecret.ResourceVersion
	}

	if a.options.GatewayConfigurationSecret != nil && k8sutils.IsSecretWatchedByApicast(a.options.GatewayConfigurationSecret) {
		annotations[GatewayConfigurationSecretResverAnnotation] = a.options.GatewayConfigurationSecret.ResourceVersion
	}

	if a.options.HTTPSCertificateSecret != nil && k8sutils.IsSecretWatchedByApicast(a.options.HTTPSCertificateSecret) {
		annotations[HttpsCertSecretResverAnnotation] = a.options.HTTPSCertificateSecret.ResourceVersion
	}

	if a.options.TracingConfig.Enabled && a.options.TracingConfig.Secret != nil && k8sutils.IsSecretWatchedByApicast(a.options.TracingConfig.Secret) {
		annotations[OpenTracingSecretResverAnnotation] = a.options.TracingConfig.Secret.ResourceVersion
	}

	if a.options.Opentelemetry.Enabled && a.options.Opentelemetry.SecretName != "" {
		telemetryConfigSecret := &v1.Secret{}
		telemetryConfigSecretKey := client.ObjectKey{
			Name:      a.options.Opentelemetry.SecretName,
			Namespace: a.options.Namespace,
		}
		err := k8sclient.Get(ctx, telemetryConfigSecretKey, telemetryConfigSecret)
		if err != nil {
			return nil, err
		}
		if k8sutils.IsSecretWatchedByApicast(telemetryConfigSecret) {
			annotations[OpenTelemetrySecretResverAnnotation] = telemetryConfigSecret.ResourceVersion
		}
	}

	for idx := range a.options.CustomEnvironments {
		// Secrets must exist and have the watched-by label
		// Annotation key includes the name of the secret
		if k8sutils.IsSecretWatchedByApicast(a.options.CustomEnvironments[idx]) {
			annotationKey := fmt.Sprintf("%s%s", CustomEnvSecretResverAnnotationPrefix, a.options.CustomEnvironments[idx].Name)
			annotations[annotationKey] = a.options.CustomEnvironments[idx].ResourceVersion
		}
	}

	for idx := range a.options.CustomPolicies {
		// Secrets must exist and have the watched-by label
		// Annotation key includes the name of the secret
		if k8sutils.IsSecretWatchedByApicast(a.options.CustomPolicies[idx].Secret) {
			annotationKey := fmt.Sprintf("%s%s", CustomPoliciesSecretResverAnnotationPrefix, a.options.CustomPolicies[idx].Secret.Name)
			annotations[annotationKey] = a.options.CustomPolicies[idx].Secret.ResourceVersion
		}
	}

	return annotations, nil
}

func (a *APIcast) podAnnotations(watchedSecretAnnotations map[string]string) map[string]string {
	annotations := map[string]string{
		"prometheus.io/scrape": "true",
		"prometheus.io/port":   "9421",
	}

	for key, val := range watchedSecretAnnotations {
		annotations[key] = val
	}

	return annotations
}

func (a *APIcast) Service() *v1.Service {
	service := &v1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      a.options.ServiceName,
			Namespace: a.options.Namespace,
			Labels:    a.options.CommonLabels,
		},
		Spec: v1.ServiceSpec{
			Ports:    a.servicePorts(),
			Selector: a.options.PodLabelSelector,
		},
	}

	addOwnerRefToObject(service, *a.options.Owner)
	return service
}

func (a *APIcast) servicePorts() []v1.ServicePort {
	servicePorts := []v1.ServicePort{
		{Name: "proxy", Port: appsv1alpha1.DefaultHTTPPort, Protocol: v1.ProtocolTCP, TargetPort: intstr.FromString("proxy")},
		{Name: "management", Port: DefaultManagementPort, Protocol: v1.ProtocolTCP, TargetPort: intstr.FromString("management")},
	}

	if a.options.HTTPSPort != nil {
		servicePorts = append(servicePorts,
			v1.ServicePort{Name: "httpsproxy", Port: *a.options.HTTPSPort, Protocol: v1.ProtocolTCP, TargetPort: intstr.FromString("httpsproxy")})
	}

	return servicePorts
}

func (a *APIcast) containerPorts() []v1.ContainerPort {
	ports := []v1.ContainerPort{
		{Name: "proxy", ContainerPort: appsv1alpha1.DefaultHTTPPort, Protocol: v1.ProtocolTCP},
		{Name: "management", ContainerPort: DefaultManagementPort, Protocol: v1.ProtocolTCP},
		{Name: "metrics", ContainerPort: DefaultMetricsPort, Protocol: v1.ProtocolTCP},
	}

	if a.options.HTTPSPort != nil {
		ports = append(ports,
			v1.ContainerPort{Name: "httpsproxy", ContainerPort: *a.options.HTTPSPort, Protocol: v1.ProtocolTCP})
	}

	return ports
}

func (a *APIcast) livenessProbe() *v1.Probe {
	return &v1.Probe{
		ProbeHandler: v1.ProbeHandler{
			HTTPGet: &v1.HTTPGetAction{
				Path: "/status/live",
				Port: intstr.FromInt32(8090),
			},
		},
		InitialDelaySeconds: 10,
		TimeoutSeconds:      5,
		PeriodSeconds:       10,
	}
}

func (a *APIcast) readinessProbe() *v1.Probe {
	return &v1.Probe{
		ProbeHandler: v1.ProbeHandler{
			HTTPGet: &v1.HTTPGetAction{
				Path: "/status/ready",
				Port: intstr.FromInt32(8090),
			},
		},
		InitialDelaySeconds: 15,
		TimeoutSeconds:      5,
		PeriodSeconds:       30,
	}
}

func (a *APIcast) Ingress() *networkingv1.Ingress {
	ingressPathType := networkingv1.PathTypeImplementationSpecific
	ingress := &networkingv1.Ingress{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "networking.k8s.io/v1",
			Kind:       "Ingress",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      a.options.DeploymentName,
			Namespace: a.options.Namespace,
			Labels:    a.options.CommonLabels,
		},
		Spec: networkingv1.IngressSpec{
			IngressClassName: a.options.ExposedHost.IngressClassName,
			TLS:              a.options.ExposedHost.TLS,
			Rules: []networkingv1.IngressRule{
				{
					Host: a.options.ExposedHost.Host,
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									PathType: &ingressPathType,
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: a.options.ServiceName,
											Port: networkingv1.ServiceBackendPort{
												Name: "proxy",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	addOwnerRefToObject(ingress, *a.options.Owner)
	return ingress
}

func (a *APIcast) HashedSecret(ctx context.Context, k8sclient client.Client, secretRefs []*v1.LocalObjectReference) (*v1.Secret, error) {
	hashedSecretData, err := a.computeHashedSecretData(ctx, k8sclient, secretRefs)
	if err != nil {
		return nil, err
	}

	return &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      HashedSecretName,
			Namespace: a.options.Namespace,
			Labels:    a.options.CommonLabels,
		},
		StringData: hashedSecretData,
		Type:       v1.SecretTypeOpaque,
	}, nil
}

func (a *APIcast) computeHashedSecretData(ctx context.Context, k8sclient client.Client, secretRefs []*v1.LocalObjectReference) (map[string]string, error) {
	data := make(map[string]string)

	for _, secretRef := range secretRefs {
		secret := &v1.Secret{}
		key := client.ObjectKey{
			Name:      secretRef.Name,
			Namespace: a.options.Namespace,
		}
		err := k8sclient.Get(ctx, key, secret)
		if err != nil {
			return nil, err
		}

		if k8sutils.IsSecretWatchedByApicast(secret) {
			data[secretRef.Name] = HashSecret(secret.Data)
		}
	}

	return data, nil
}

func HashSecret(data map[string][]byte) string {
	hash := sha256.New()

	sortedKeys := make([]string, 0, len(data))
	for k := range data {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)

	for _, key := range sortedKeys {
		value := data[key]
		combinedKeyValue := append([]byte(key), value...)
		hash.Write(combinedKeyValue)
	}

	hashBytes := hash.Sum(nil)

	return hex.EncodeToString(hashBytes)
}

func (a *APIcast) hasSecretHashChanged(ctx context.Context, k8sclient client.Client, deploymentAnnotation string, hashedSecret *v1.Secret) bool {
	logger, _ := logr.FromContext(ctx)

	secretToCheck := &v1.Secret{}
	secretToCheckKey := client.ObjectKey{
		Namespace: a.options.Namespace,
	}

	// Assign the name of the secret to check
	switch {
	case deploymentAnnotation == AdmPortalSecretResverAnnotation:
		secretToCheckKey.Name = a.options.AdminPortalCredentialsSecret.Name
	case deploymentAnnotation == GatewayConfigurationSecretResverAnnotation:
		secretToCheckKey.Name = a.options.GatewayConfigurationSecret.Name
	case deploymentAnnotation == HttpsCertSecretResverAnnotation:
		secretToCheckKey.Name = a.options.HTTPSCertificateSecret.Name
	case deploymentAnnotation == OpenTracingSecretResverAnnotation:
		secretToCheckKey.Name = a.options.TracingConfig.Secret.Name
	case deploymentAnnotation == OpenTelemetrySecretResverAnnotation:
		secretToCheckKey.Name = a.options.Opentelemetry.SecretName
	case strings.HasPrefix(deploymentAnnotation, CustomEnvSecretResverAnnotationPrefix):
		secretToCheckKey.Name = strings.TrimPrefix(deploymentAnnotation, CustomEnvSecretResverAnnotationPrefix)
	case strings.HasPrefix(deploymentAnnotation, CustomPoliciesSecretResverAnnotationPrefix):
		secretToCheckKey.Name = strings.TrimPrefix(deploymentAnnotation, CustomPoliciesSecretResverAnnotationPrefix)
	default:
		return false
	}

	// Get latest version of the secret to check
	err := k8sclient.Get(ctx, secretToCheckKey, secretToCheck)
	if err != nil {
		logger.Error(err, fmt.Sprintf("failed to get secret %s", secretToCheckKey.Name))
		return false
	}

	// Compare the hash of the latest version of the secret's data to the reference in the hashed secret
	if HashSecret(secretToCheck.Data) != k8sutils.SecretStringDataFromData(hashedSecret)[secretToCheckKey.Name] {
		logger.V(1).Info(fmt.Sprintf("%s secret .data has changed - updating the resourceVersion in deployment's annotation", secretToCheckKey.Name))
		return true
	}

	logger.V(1).Info(fmt.Sprintf("%s secret .data has not changed since last checked", secretToCheckKey.Name))
	return false
}

func addOwnerRefToObject(o metav1.Object, owner metav1.OwnerReference) {
	o.SetOwnerReferences(append(o.GetOwnerReferences(), owner))
}
