#!/bin/bash

set -ev

# Download kubectl
echo "######### Download Kubectl #########"
curl -Lo kubectl https://storage.googleapis.com/kubernetes-release/release/v1.20.1/bin/linux/ppc64le/kubectl && chmod +x kubectl && sudo mv kubectl /usr/local/bin/

# Download minikube.
echo "######### Download Minikube #########"
curl -Lo minikube https://storage.googleapis.com/minikube/releases/v1.20.0/minikube-linux-ppc64le && chmod +x minikube && sudo mv minikube /usr/local/bin/
mkdir -p $HOME/.kube $HOME/.minikube
touch $KUBECONFIG


# Download OPERATOR_SDK
echo "######### Download Operator SDK #########"
export OPERATOR_SDK_RELEASE_VERSION=v1.2.0
curl -OJL https://github.com/operator-framework/operator-sdk/releases/download/${OPERATOR_SDK_RELEASE_VERSION}/operator-sdk-${OPERATOR_SDK_RELEASE_VERSION}-ppc64le-linux-gnu
sudo mv operator-sdk-${OPERATOR_SDK_RELEASE_VERSION}-ppc64le-linux-gnu /usr/local/bin/operator-sdk && chmod +x /usr/local/bin/operator-sdk

# Start minikube
echo "######### Starting Minikube #########"
sudo -E minikube start --force --profile=minikube --driver=none
minikube update-context --profile=minikube
