#!/bin/bash

set -ev

echo "######### Bundle Validate  #########"
make bundle

echo "######### Lint #########"
make download
make run-lint

echo "######### Test CRDs #########"
make download
make test-crds

echo "######### Tests Manifests Version #########"
make download
make test-manifests-version

echo "######### Run e2e Test #########"
make download
make test-e2e

echo "######### License Check #########"
gem install license_finder --version 5.7.1
make licenses-check
