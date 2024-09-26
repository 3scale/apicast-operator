# Setting SHELL to bash allows bash commands to be executed by recipes.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

MKFILE_PATH := $(abspath $(lastword $(MAKEFILE_LIST)))
PROJECT_PATH := $(patsubst %/,%,$(dir $(MKFILE_PATH)))

all: manager

# Current Operator version
VERSION ?= 0.8.0
# Default bundle image tag
BUNDLE_IMG ?= controller-bundle:$(VERSION)
# Options for 'bundle-build'
ifneq ($(origin CHANNELS), undefined)
BUNDLE_CHANNELS := --channels=$(CHANNELS)
endif
ifneq ($(origin DEFAULT_CHANNEL), undefined)
BUNDLE_DEFAULT_CHANNEL := --default-channel=$(DEFAULT_CHANNEL)
endif
BUNDLE_METADATA_OPTS ?= $(BUNDLE_CHANNELS) $(BUNDLE_DEFAULT_CHANNEL)

# Image URL to use all building/pushing image targets
IMG ?= quay.io/3scale/apicast-operator:master

CRD_OPTIONS ?= "crd:crdVersions=v1"

DOCKER ?= docker
KUBECTL ?= kubectl

OS := $(shell uname | awk '{print tolower($$0)}' | sed -e s/linux/linux-gnu/ )
ARCH := $(shell uname -m)

LICENSEFINDERBINARY := $(shell command -v license_finder 2> /dev/null)
DEPENDENCY_DECISION_FILE = $(PROJECT_PATH)/doc/dependency_decisions.yml

NAMESPACE ?= $(shell $(KUBECTL) config view --minify -o jsonpath='{.contexts[0].context.namespace}' 2>/dev/null || echo operator-test)

CURRENT_DATE=$(shell date +%s)

# find or download controller-gen
# download controller-gen if necessary
CONTROLLER_GEN=$(PROJECT_PATH)/bin/controller-gen
$(CONTROLLER_GEN):
	$(call go-bin-install,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen@v0.9.2)

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN)

KUSTOMIZE=$(PROJECT_PATH)/bin/kustomize
$(KUSTOMIZE):
	$(call go-bin-install,$(KUSTOMIZE),sigs.k8s.io/kustomize/kustomize/v4@v4.5.7)

.PHONY: kustomize
kustomize: $(KUSTOMIZE)

OPERATOR_SDK = $(PROJECT_PATH)/bin/operator-sdk
# Note: release file patterns changed after v1.2.0
# More info https://sdk.operatorframework.io/docs/installation/
OPERATOR_SDK_VERSION=v1.2.0
$(OPERATOR_SDK):
	mkdir -p $(PROJECT_PATH)/bin
	curl -sSL https://github.com/operator-framework/operator-sdk/releases/download/$(OPERATOR_SDK_VERSION)/operator-sdk-${OPERATOR_SDK_VERSION}-$(ARCH)-${OS} -o $(OPERATOR_SDK)
	chmod +x $(OPERATOR_SDK)

.PHONY: operator-sdk
operator-sdk: $(OPERATOR_SDK)

# find or download yq
YQ=$(PROJECT_PATH)/bin/yq
$(YQ):
	$(call go-bin-install,$(YQ),github.com/mikefarah/yq/v4@latest)

.PHONY: yq
yq: $(YQ)

# Run all tests
test: test-unit test-integration

# Build manager binary
manager: generate fmt vet
	go build -o bin/manager main.go

# Run against the configured Kubernetes cluster in ~/.kube/config
run: export WATCH_NAMESPACE=$(NAMESPACE)
run: generate fmt vet manifests
	go run ./main.go --zap-devel

# Install CRDs into a cluster
install: manifests $(KUSTOMIZE)
	$(KUSTOMIZE) build config/crd | $(KUBECTL) apply -f -

# Uninstall CRDs from a cluster
uninstall: manifests $(KUSTOMIZE)
	$(KUSTOMIZE) build config/crd | $(KUBECTL) delete -f -

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: manifests $(KUSTOMIZE)
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default | $(KUBECTL) apply -f -

# Generate manifests e.g. CRD, RBAC etc.
manifests: $(CONTROLLER_GEN)
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

# Generate code
generate: $(CONTROLLER_GEN)
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

# Build the docker image
.PHONY: docker-build
docker-build: test docker-build-only

.PHONY: docker-build-only
docker-build-only:
	$(DOCKER) build . -t ${IMG}

# Push the operator docker image
.PHONY: operator-image-push
operator-image-push:
	$(DOCKER) push ${IMG}

# Push the bundle docker image
.PHONY: bundle-image-push
bundle-image-push:
	$(DOCKER) push ${BUNDLE_IMG}

# Generate bundle manifests and metadata, then validate generated files.
.PHONY: bundle
bundle: manifests $(KUSTOMIZE) $(OPERATOR_SDK)
	$(OPERATOR_SDK) generate kustomize manifests -q
	cd config/manager && $(KUSTOMIZE) edit set image controller=$(IMG)
	$(KUSTOMIZE) build config/manifests | $(OPERATOR_SDK) generate bundle -q --overwrite --version $(VERSION) $(BUNDLE_METADATA_OPTS)
	$(OPERATOR_SDK) bundle validate ./bundle

.PHONY: bundle-update-test
bundle-update-test:
	git diff --exit-code ./bundle
	[ -z "$$(git ls-files --other --exclude-standard --directory --no-empty-directory ./bundle)" ]

# Build the bundle image.
.PHONY: bundle-build
bundle-build: bundle-validate
	$(DOCKER) build -f bundle.Dockerfile -t $(BUNDLE_IMG) .

## 3scale-specific targets

download:
	@echo Download go.mod dependencies
	@go mod download

## licenses.xml: Generate licenses.xml file
licenses.xml:
ifndef LICENSEFINDERBINARY
	$(error "license-finder is not available please install: gem install license_finder --version 5.7.1")
endif
	license_finder report --decisions-file=$(DEPENDENCY_DECISION_FILE) --quiet --format=xml > licenses.xml

## licenses-check: Check license compliance of dependencies
licenses-check:
ifndef LICENSEFINDERBINARY
	$(error "license-finder is not available please install: gem install license_finder --version 5.7.1")
endif
	@echo "Checking license compliance"
	license_finder --decisions-file=$(DEPENDENCY_DECISION_FILE)

# Run unit tests
test-unit: generate fmt vet manifests ## Run Unit tests.
	go test ./... -tags unit -v -timeout 0

# Run integration tests
# Runnning integration tests with a local kind cluster until envtest issues are fixed
# 1) make test does not terminate kube-apiserver and etcd processes https://github.com/kubernetes-sigs/cluster-api-provider-aws/issues/2753
# 2) Namespace usage limitation https://book.kubebuilder.io/reference/envtest.html#namespace-usage-limitation
# 3) Timeout waiting for process kube-apiserver to stop https://github.com/kubernetes-sigs/controller-runtime/issues/1571
test-integration: export USE_EXISTING_CLUSTER=true
test-integration: generate fmt vet manifests ## Run Integration tests.
	go test ./... -tags integration -ginkgo.v -v -timeout 600s

.PHONY: bundle-validate
bundle-validate: $(OPERATOR_SDK)
	$(OPERATOR_SDK) bundle validate ./bundle

.PHONY: bundle-validate-image
bundle-validate-image: $(OPERATOR_SDK)
	$(OPERATOR_SDK) bundle validate $(BUNDLE_IMG)

.PHONY: bundle-custom-updates
bundle-custom-updates: BUNDLE_PREFIX=dev$(CURRENT_DATE)
bundle-custom-updates: $(YQ)
	@echo "Update metadata to avoid collision with existing APIcast Operator official public operators catalog entries"
	@echo "using BUNDLE_PREFIX $(BUNDLE_PREFIX)"
	$(YQ) --inplace '.metadata.name = "$(BUNDLE_PREFIX)-apicast-operator.$(VERSION)"' $(PROJECT_PATH)/bundle/manifests/apicast-operator.clusterserviceversion.yaml
	$(YQ) --inplace '.spec.displayName = "$(BUNDLE_PREFIX) apicast operator"' $(PROJECT_PATH)/bundle/manifests/apicast-operator.clusterserviceversion.yaml
	$(YQ) --inplace '.spec.provider.name = "$(BUNDLE_PREFIX)"' $(PROJECT_PATH)/bundle/manifests/apicast-operator.clusterserviceversion.yaml
	$(YQ) --inplace '.annotations."operators.operatorframework.io.bundle.package.v1" = "$(BUNDLE_PREFIX)-apicast-operator"' $(PROJECT_PATH)/bundle/metadata/annotations.yaml
	sed -E -i 's/(operators\.operatorframework\.io\.bundle\.package\.v1=).+/\1$(BUNDLE_PREFIX)-apicast-operator/' $(PROJECT_PATH)/bundle.Dockerfile
	@echo "Update operator image reference URL"
	$(YQ) --inplace '.metadata.annotations.containerImage = "$(IMG)"' $(PROJECT_PATH)/bundle/manifests/apicast-operator.clusterserviceversion.yaml
	$(YQ) --inplace '.spec.install.spec.deployments[0].spec.template.spec.containers[0].image = "$(IMG)"' $(PROJECT_PATH)/bundle/manifests/apicast-operator.clusterserviceversion.yaml

.PHONY: bundle-restore
bundle-restore:
	git checkout bundle/manifests/apicast-operator.clusterserviceversion.yaml bundle/metadata/annotations.yaml bundle.Dockerfile

.PHONY: bundle-custom-build
bundle-custom-build: | bundle-custom-updates bundle-build bundle-restore

.PHONY: bundle-run
bundle-run: $(OPERATOR_SDK)
	$(OPERATOR_SDK) run bundle --namespace openshift-marketplace $(BUNDLE_IMG)

GOLANGCI-LINT=$(PROJECT_PATH)/bin/golangci-lint
$(GOLANGCI-LINT):
	mkdir -p $(PROJECT_PATH)/bin
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(PROJECT_PATH)/bin v1.52.2

.PHONY: golangci-lint
golangci-lint: $(GOLANGCI-LINT)

.PHONY: run-lint
run-lint: $(GOLANGCI-LINT)
	$(GOLANGCI-LINT) run --timeout 5m

# Include last to avoid changing MAKEFILE_LIST used above
include ./make/*.mk
