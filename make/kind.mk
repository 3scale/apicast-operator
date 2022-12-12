
##@ Kind

## Targets to help install and use kind for development https://kind.sigs.k8s.io

KIND = $(PROJECT_PATH)/bin/kind
$(KIND): ## Download kind locally if necessary.
	$(call go-bin-install,$(KIND),sigs.k8s.io/kind@v0.11.1)

.PHONY: kind
kind: $(KIND)

KIND_CLUSTER_NAME ?= threescale-local

.PHONY: kind-create-cluster
kind-create-cluster: kind ## Create the "kuadrant-local" kind cluster.
	$(KIND) create cluster --name $(KIND_CLUSTER_NAME) --config utils/kind-cluster.yaml

.PHONY: kind-delete-cluster
kind-delete-cluster: kind ## Delete the "kuadrant-local" kind cluster.
	- $(KIND) delete cluster --name $(KIND_CLUSTER_NAME)
