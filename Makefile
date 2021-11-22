SHELL := /bin/bash

include utils.mk

# Current Operator version
VERSION ?= 0.0.1
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

MKFILE_PATH := $(abspath $(lastword $(MAKEFILE_LIST)))
PROJECT_PATH := $(patsubst %/,%,$(dir $(MKFILE_PATH)))

OS := $(shell uname | awk '{print tolower($$0)}' | sed -e s/linux/linux-gnu/ )
ARCH := $(shell uname -m)

LICENSEFINDERBINARY := $(shell command -v license_finder 2> /dev/null)
DEPENDENCY_DECISION_FILE = $(PROJECT_PATH)/doc/dependency_decisions.yml

NAMESPACE ?= $(shell $(KUBECTL) config view --minify -o jsonpath='{.contexts[0].context.namespace}' 2>/dev/null || echo operator-test)

CURRENT_DATE=$(shell date +%s)

all: manager

# find or download controller-gen
# download controller-gen if necessary
CONTROLLER_GEN=$(PROJECT_PATH)/bin/controller-gen
$(CONTROLLER_GEN):
	$(call go-get-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen@v0.3.0)

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN)

KUSTOMIZE=$(PROJECT_PATH)/bin/kustomize
$(KUSTOMIZE):
	$(call go-get-tool,$(KUSTOMIZE),sigs.k8s.io/kustomize/kustomize/v3@v3.5.4)

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
	$(call go-get-tool,$(YQ),github.com/mikefarah/yq/v3)

.PHONY: yq
yq: $(YQ)

# Run all tests
test: test-unit test-crds test-e2e

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
TEST_UNIT_PKGS = $(shell go list ./... | grep -E 'github.com/3scale/apicast-operator/pkg/|github.com/3scale/apicast-operator/pkg/apis')
test-unit: generate fmt vet manifests
	go test -v $(TEST_UNIT_PKGS)

# Run CRD tests
TEST_CRD_PKGS = $(shell go list ./... | grep 'github.com/3scale/apicast-operator/test/crds')
test-crds: generate fmt vet manifests
	go test -v $(TEST_CRD_PKGS)

# Run e2e tests

TEST_E2E_PKGS = $(shell go list ./... | grep -E 'github.com/3scale/apicast-operator/controllers')
ENVTEST_ASSETS_DIR=$(PROJECT_PATH)/testbin
test-e2e: generate fmt vet manifests
	mkdir -p ${ENVTEST_ASSETS_DIR}
	test -f $(ENVTEST_ASSETS_DIR)/setup-envtest.sh || curl -sSLo $(ENVTEST_ASSETS_DIR)/setup-envtest.sh https://raw.githubusercontent.com/kubernetes-sigs/controller-runtime/v0.8.0/hack/setup-envtest.sh
	source ${ENVTEST_ASSETS_DIR}/setup-envtest.sh; fetch_envtest_tools $(ENVTEST_ASSETS_DIR); setup_envtest_env $(ENVTEST_ASSETS_DIR); go test $(TEST_E2E_PKGS) -coverprofile cover.out -ginkgo.v -ginkgo.progress -v

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
	$(YQ) w --inplace $(PROJECT_PATH)/bundle/manifests/apicast-operator.clusterserviceversion.yaml metadata.name $(BUNDLE_PREFIX)-apicast-operator.$(VERSION)
	$(YQ) w --inplace $(PROJECT_PATH)/bundle/manifests/apicast-operator.clusterserviceversion.yaml spec.displayName "$(BUNDLE_PREFIX) apicast operator"
	$(YQ) w --inplace $(PROJECT_PATH)/bundle/manifests/apicast-operator.clusterserviceversion.yaml spec.provider.name $(BUNDLE_PREFIX)
	$(YQ) w --inplace $(PROJECT_PATH)/bundle/metadata/annotations.yaml 'annotations."operators.operatorframework.io.bundle.package.v1"' $(BUNDLE_PREFIX)-apicast-operator
	sed -E -i 's/(operators\.operatorframework\.io\.bundle\.package\.v1=).+/\1$(BUNDLE_PREFIX)-apicast-operator/' $(PROJECT_PATH)/bundle.Dockerfile
	@echo "Update operator image reference URL"
	$(YQ) w --inplace $(PROJECT_PATH)/bundle/manifests/apicast-operator.clusterserviceversion.yaml metadata.annotations.containerImage $(IMG)
	$(YQ) w --inplace $(PROJECT_PATH)/bundle/manifests/apicast-operator.clusterserviceversion.yaml spec.install.spec.deployments[0].spec.template.spec.containers[0].image $(IMG)

.PHONY: bundle-restore
bundle-restore:
	git checkout bundle/manifests/apicast-operator.clusterserviceversion.yaml bundle/metadata/annotations.yaml bundle.Dockerfile

.PHONY: bundle-custom-build
bundle-custom-build: | bundle-custom-updates bundle-build bundle-restore

.PHONY: bundle-run
bundle-run: $(OPERATOR_SDK)
	$(OPERATOR_SDK) run bundle --namespace $(NAMESPACE) $(BUNDLE_IMG)

GOLANGCI-LINT=$(PROJECT_PATH)/bin/golangci-lint
$(GOLANGCI-LINT):
	mkdir -p $(PROJECT_PATH)/bin
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(PROJECT_PATH)/bin v1.41.1

.PHONY: golangci-lint
golangci-lint: $(GOLANGCI-LINT)

.PHONY: run-lint
run-lint: $(GOLANGCI-LINT)
	$(GOLANGCI-LINT) run
