SHELL := /bin/bash
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
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true"

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

OPERATOR_SDK ?= operator-sdk
DOCKER ?= docker
KUBECTL ?= kubectl
YQ := $(shell command -v yq 2> /dev/null)

MKFILE_PATH := $(abspath $(lastword $(MAKEFILE_LIST)))
PROJECT_PATH := $(patsubst %/,%,$(dir $(MKFILE_PATH)))

LICENSEFINDERBINARY := $(shell command -v license_finder 2> /dev/null)
DEPENDENCY_DECISION_FILE = $(PROJECT_PATH)/doc/dependency_decisions.yml

NAMESPACE ?= $(shell $(KUBECTL) config view --minify -o jsonpath='{.contexts[0].context.namespace}' 2>/dev/null || echo operator-test)

CURRENT_DATE=$(shell date +%s)

all: manager

# Run all tests
test: test-unit test-crds test-e2e

# Build manager binary
manager: generate fmt vet
	go build -o bin/manager main.go

# Run against the configured Kubernetes cluster in ~/.kube/config
run: export WATCH_NAMESPACE=$(NAMESPACE)
run: generate fmt vet manifests
	go run ./main.go

# Install CRDs into a cluster
install: manifests kustomize
	$(KUSTOMIZE) build config/crd | $(KUBECTL) apply -f -

# Uninstall CRDs from a cluster
uninstall: manifests kustomize
	$(KUSTOMIZE) build config/crd | $(KUBECTL) delete -f -

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: manifests kustomize
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default | $(KUBECTL) apply -f -

# Generate manifests e.g. CRD, RBAC etc.
manifests: controller-gen
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

# Generate code
generate: controller-gen
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

# Build the docker image
docker-build: test
	$(DOCKER) build . -t ${IMG}

docker-build-only:
	$(DOCKER) build . -t ${IMG}

# Push the operator docker image
operator-docker-image-push:
	$(DOCKER) push ${IMG}

# Push the bundle docker image
bundle-docker-image-push:
	$(DOCKER) push ${BUNDLE_IMG}

# find or download controller-gen
# download controller-gen if necessary
controller-gen:
ifeq (, $(shell which controller-gen))
	@{ \
	set -e ;\
	CONTROLLER_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$CONTROLLER_GEN_TMP_DIR ;\
	go mod init tmp ;\
	go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.3.0 ;\
	rm -rf $$CONTROLLER_GEN_TMP_DIR ;\
	}
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif

kustomize:
ifeq (, $(shell which kustomize))
	@{ \
	set -e ;\
	KUSTOMIZE_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$KUSTOMIZE_GEN_TMP_DIR ;\
	go mod init tmp ;\
	go get sigs.k8s.io/kustomize/kustomize/v3@v3.5.4 ;\
	rm -rf $$KUSTOMIZE_GEN_TMP_DIR ;\
	}
KUSTOMIZE=$(GOBIN)/kustomize
else
KUSTOMIZE=$(shell which kustomize)
endif

# Generate bundle manifests and metadata, then validate generated files.
.PHONY: bundle
bundle: manifests kustomize
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
	test -f $(ENVTEST_ASSETS_DIR)/setup-envtest.sh || curl -sSLo $(ENVTEST_ASSETS_DIR)/setup-envtest.sh https://raw.githubusercontent.com/kubernetes-sigs/controller-runtime/v0.6.3/hack/setup-envtest.sh
	source ${ENVTEST_ASSETS_DIR}/setup-envtest.sh; fetch_envtest_tools $(ENVTEST_ASSETS_DIR); setup_envtest_env $(ENVTEST_ASSETS_DIR); go test $(TEST_E2E_PKGS) -coverprofile cover.out -ginkgo.v -ginkgo.progress -v

.PHONY: bundle-validate
bundle-validate:
	$(OPERATOR_SDK) bundle validate ./bundle

.PHONY: bundle-validate-image
bundle-validate-image:
	$(OPERATOR_SDK) bundle validate $(BUNDLE_IMG)

.PHONY: bundle-custom-updates
bundle-custom-updates: BUNDLE_PREFIX=dev$(CURRENT_DATE)
bundle-custom-updates:
ifndef YQ
	$(error "yq is not available please install: https://github.com/mikefarah/yq/releases/latest")
endif
	@echo "Update metadata to avoid collision with existing APIcast Operator official public operators catalog entries"
	$(YQ) w --inplace $(PROJECT_PATH)/bundle/manifests/apicast-operator.clusterserviceversion.yaml metadata.name $(BUNDLE_PREFIX)-apicast-operator.v0.0.1
	$(YQ) w --inplace $(PROJECT_PATH)/bundle/manifests/apicast-operator.clusterserviceversion.yaml spec.displayName "$(BUNDLE_PREFIX) apicast operator"
	$(YQ) w --inplace $(PROJECT_PATH)/bundle/manifests/apicast-operator.clusterserviceversion.yaml spec.provider.name $(BUNDLE_PREFIX)
	$(YQ) w --inplace $(PROJECT_PATH)/bundle/metadata/annotations.yaml 'annotations."operators.operatorframework.io.bundle.package.v1"' $(BUNDLE_PREFIX)-apicast-operator
	sed -E -i 's/(operators\.operatorframework\.io\.bundle\.package\.v1=).+/\1$(BUNDLE_PREFIX)-apicast-operator/' $(PROJECT_PATH)/bundle.Dockerfile
	@echo "Update operator image reference URL"
	$(YQ) w --inplace $(PROJECT_PATH)/bundle/manifests/apicast-operator.clusterserviceversion.yaml metadata.annotations.containerImage $(IMG)
	$(YQ) w --inplace $(PROJECT_PATH)/bundle/manifests/apicast-operator.clusterserviceversion.yaml spec.install.spec.deployments[0].spec.template.spec.containers[1].image $(IMG)

.PHONY: bundle-restore
bundle-restore:
	git checkout bundle/manifests/apicast-operator.clusterserviceversion.yaml bundle/metadata/annotations.yaml bundle.Dockerfile

.PHONY: bundle-custom-build
bundle-custom-build: | bundle-custom-updates bundle-build bundle-restore

.PHONY: bundle-run
bundle-run:
	$(OPERATOR_SDK) run bundle --namespace $(NAMESPACE) $(BUNDLE_IMG)
